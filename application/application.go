package application

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/GehirnInc/crypt"
	"github.com/dedalqq/omg.httpserver"
	"github.com/pkg/errors"

	_ "github.com/GehirnInc/crypt/md5_crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"

	"qubert/logger"
	"qubert/resources"

	"qubert/plugins/interfaces"
	"qubert/plugins/services"
	"qubert/plugins/system"
	"qubert/plugins/systemd"
)

type Application struct {
	cfg *Config
	log *logger.Logger
	pc  *pluginController
	sm  *sessionManager

	version string
	commit  string
}

func NewApplication(cfg *Config, version string, commit string) *Application {
	return &Application{
		cfg:     cfg,
		log:     logger.CreateLogger(cfg.Debug),
		version: version,
		commit:  commit,
	}
}

func (a *Application) authUser(login, password string) (ok bool, err error) {
	shadowData, err := ioutil.ReadFile("/etc/shadow")
	if err != nil {
		return false, err
	}

	r := bufio.NewReader(bytes.NewReader(shadowData))

	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("Panic: %v", r)
		}
	}()

	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				return false, nil
			}

			return false, err
		}

		lineParts := strings.Split(string(line), ":")

		if lineParts[0] != login {
			continue
		}

		err = crypt.NewFromHash(lineParts[1]).Verify(lineParts[1], []byte(password))
		if err != nil {
			if err == crypt.ErrKeyMismatch {
				continue
			}

			return false, err
		}

		return true, nil
	}
}

func getRouter(a *Application, rs *resources.Storage) httpserver.Router {
	r := httpserver.NewRouter()

	r.Default(newDefaultHandler(rs))

	apiSubRoute := r.SubRoute("/api")

	apiSubRoute.Add("/user", newUserHandler(a))
	apiSubRoute.Add("/login", newLoginHandler(a))
	apiSubRoute.Add("/main", newMainPageHandler(a))
	apiSubRoute.Add("/plugins/{any}", newPluginRenderHandler(a))
	apiSubRoute.Add("/plugins/{any}/action", newPluginActionHandler(a))

	r.Add("/ws", newWebSocketHandler(a))

	return r
}

func (a *Application) Run(ctx context.Context) error {
	a.log.Info("Init...")

	a.sm = newSessionManager(ctx)

	a.pc = newPluginController(
		a.cfg.SettingsFile,
		a.sm,
		&services.Plugin{},
		&interfaces.Plugin{},
		&systemd.Plugin{},
		&system.Plugin{},
	)

	a.pc.setVersion(a.version, a.commit)

	ctx = makeContext(ctx, a.log)

	rs := resources.NewStorage()

	router := getRouter(a, rs)

	server := httpserver.NewServer(
		ctx,
		fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port),
		router,
		httpserver.Options{
			SupportGZIP: true,
			Logger:      a.log,
		},
	)

	var wg sync.WaitGroup

	err := a.pc.initPlugins(ctx, &wg)
	if err != nil {
		a.log.Error(err)

		cancelContext(ctx)
	}

	runSignalHandler(ctx, &wg, a.log)
	runServer(ctx, &wg, server, a.log)
	runWaitingContext(ctx, &wg, server, a.log)

	wg.Wait()

	return nil
}

func runSignalHandler(ctx context.Context, wg *sync.WaitGroup, logger *logger.Logger) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case s := <-sigs:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					logger.Info("interrupt [%d]", s)
					cancelContext(ctx)

					return
				}
			case <-ctx.Done():
				return
			}

		}
	}()
}

func runServer(ctx context.Context, wg *sync.WaitGroup, server *http.Server, logger *logger.Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("Starting listening on [%s]", server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error(err)
			cancelContext(ctx)

			return
		}
	}()
}

func runWaitingContext(ctx context.Context, wg *sync.WaitGroup, server *http.Server, logger *logger.Logger) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		logger.Info("Context cancel, stop server")

		err := server.Close()
		if err != nil {
			logger.Error(err)
		}
	}()
}
