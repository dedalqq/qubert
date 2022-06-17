package application

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dedalqq/omg.httpserver"

	"qubert/internal/logger"
	"qubert/plugins/interfaces"
	"qubert/plugins/services"
	"qubert/plugins/system"
	"qubert/plugins/systemd"
	"qubert/resources"
)

type Application struct {
	cfg *Config
	log *logger.Logger
	pc  *pluginController
	us  *userManager
	sm  *sessionManager

	version string
	commit  string
}

func NewApplication(cfg *Config, version string, commit string) *Application {
	return &Application{
		cfg:     cfg,
		log:     logger.CreateLogger(cfg.Debug),
		us:      NewUserManager(),
		version: version,
		commit:  commit,
	}
}

func (a *Application) authUser(login, password string) (bool, error) {
	user, err := a.us.getUserByUserName(login)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.verifyUserPassword(password)
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

	var err error

	a.pc, err = newPluginController(a.cfg.SettingsFile, a.sm)
	if err != nil {
		return err
	}

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

	err = a.pc.initPlugins(ctx, &wg, a.log,
		&services.Plugin{},
		&interfaces.Plugin{},
		&systemd.Plugin{},
		&system.Plugin{},
	)

	if err != nil {
		a.log.Error(err)

		cancelContext(ctx)
	}

	extPlugins, err := a.pc.loadExternalPlugins(a.cfg.PluginDir, a.log)
	if err != nil {
		a.log.Error(err)
	}

	err = a.pc.initPlugins(ctx, &wg, a.log, extPlugins...)
	if err != nil {
		a.log.Error(err)
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
