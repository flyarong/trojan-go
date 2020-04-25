package cert

import (
	"flag"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

type certOption struct {
	mode     *string
	httpPort *string
	tlsPort  *string
	common.OptionHandler
}

func (*certOption) Name() string {
	return "cert"
}

func (*certOption) Priority() int {
	return 10
}

func (c *certOption) Handle() error {
	switch *c.mode {
	case "request":
		tlsPort = *c.tlsPort
		httpPort = *c.httpPort
		RequestCertGuide()
		return nil
	case "renew":
		tlsPort = *c.tlsPort
		httpPort = *c.httpPort
		RenewCertGuide()
		return nil
	case "INVALID":
		return common.NewError("not specified")
	default:
		err := common.NewError("invalid args " + *c.mode)
		log.Error(err)
		return common.NewError("invalid args")
	}
}

func init() {
	common.RegisterOptionHandler(&certOption{
		mode:     flag.String("autocert", "INVALID", "Simple letsencrpyt cert ACME client. Use \"-autocert request\" to request a cert or \"-autocert renew\" to renew a cert"),
		tlsPort:  flag.String("autocert-tls-port", "443", "autocert TLS acme challenge port"),
		httpPort: flag.String("autocert-http-port", "80", "autocert HTTP acme challenge port"),
	})
}
