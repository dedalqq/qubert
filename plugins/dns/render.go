package dns

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

func updateZone(filePath string, zone *Zone) error {
	data := make([]byte, 0, 2*1024)
	buf := bytes.NewBuffer(data)

	err := renderZone(buf, zone)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(filePath, zone.Origin), buf.Bytes(), 0644)
}

func renderZone(f io.Writer, zone *Zone) error {
	err := fprintf(f,
		line("$ORIGIN %s.", zone.Origin),
		line("$TTL %s", duration(zone.TTL)),
		emptyLine(),
	)
	if err != nil {
		return err
	}

	var nameServer []string

	for _, n := range zone.NameServers {
		nameServer = append(nameServer, n.Name)
	}

	err = fprintf(f,
		line("@       IN      SOA     %s (", strings.Join(nameServer, " ")),
		line("    %d ; Serial", zone.SOA.Serial),
		line("    %s ; Refresh", duration(zone.SOA.Refresh)),
		line("    %s ; Retry", duration(zone.SOA.Retry)),
		line("    %s ; Expire", duration(zone.SOA.Expire)),
		line("    %s ; Minimum TTL", duration(zone.SOA.MinimumTTL)),
		line(")"),
		emptyLine(),
	)
	if err != nil {
		return err
	}

	for _, n := range zone.NameServers {
		err = fprintf(f, line("    IN      NS      %s", n.Name))
		if err != nil {
			return err
		}
	}

	err = fprintf(f, emptyLine())
	if err != nil {
		return err
	}

	for _, n := range zone.NameServers {
		err = fprintf(f, line("%s IN A %s", n.Name, n.Addr.String()))
		if err != nil {
			return err
		}
	}

	err = fprintf(f, emptyLine(), line("; ###"), emptyLine())
	if err != nil {
		return err
	}

	for _, r := range zone.Records {
		if r.Type == RecordTypeMX {
			err = fprintf(f, line("%s IN %s %d %s", r.Name, r.Type, r.Priority, r.Value))
		} else {
			err = fprintf(f, line("%s IN %s %s", r.Name, r.Type, r.Value))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func renderZoneDefinition(f io.Writer, dir string, zone *Zone) error {
	return fprintf(f,
		line("zone \"%s\" IN {", zone.ZoneName()),
		line("    type master;"),
		line("    file \"%s/%s\";", dir, zone.ZoneName()),
		line("    allow-query { any; };"),
		line("    allow-transfer { none; };"),
		line("    allow-update { none; };"),
		line("};"),
		emptyLine(),
	)
}

func updateConfig(filePath string, dir string, zones []*Zone) error {
	data := make([]byte, 0, 512)
	buf := bytes.NewBuffer(data)

	for _, z := range zones {
		err := renderZoneDefinition(buf, dir, z)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

type printHandler func(io.Writer) error

func line(text string, a ...any) printHandler {
	return func(writer io.Writer) error {
		_, err := fmt.Fprintf(writer, text, a...)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "\n")
		return err
	}
}

func emptyLine() printHandler {
	return func(writer io.Writer) error {
		_, err := fmt.Fprintf(writer, "\n")
		return err
	}
}

func fprintf(writer io.Writer, handlers ...printHandler) error {
	for _, h := range handlers {
		err := h(writer)
		if err != nil {
			return err
		}
	}

	return nil
}

func duration(d time.Duration) string {
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
