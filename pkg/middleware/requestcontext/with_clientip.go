package requestcontext

import (
	"context"
	"log/slog"
	"net"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type clientIPKey struct{}

type WithClientIPConfig struct {
	// [Optional] TrustedProxiesIP is a list of all proxies IP ranges that's between the server and the client.
	//
	// If it's provided, it will walk backwards from the last IP in `X-Forwarded-For` header
	// and use first IP that's not trusted proxy(not in the given IP ranges.)
	//
	// **If you want to use this option, you should provide all of probable proxies IP ranges.**
	//
	// This is lowest priority.
	TrustedProxiesIP []string `env:"TRUSTED_PROXIES_IP" mapstructure:"trusted_proxies_ip"`

	// [Optional] TrustedHeader is a header name for getting client IP. (e.g. X-Real-IP, CF-Connecting-IP, etc.)
	//
	// This is highest priority, it will ignore rest of the options if it's provided.
	TrustedHeader string `env:"TRUSTED_HEADER" mapstructure:"trusted_proxies_header"`

	// EnableRejectMalformedRequest return 403 Forbidden if the request is from proxies, but can't extract client IP
	EnableRejectMalformedRequest bool `env:"ENABLE_REJECT_MALFORMED_REQUEST" envDefault:"false" mapstructure:"enable_reject_malformed_request"`
}

// WithClientIP setup client IP context with XFF Spoofing prevention support.
//
// If request is from proxies, it will use first IP from `X-Forwarded-For` header by default.
func WithClientIP(config WithClientIPConfig) Option {
	var trustedProxies trustedProxy
	if len(config.TrustedProxiesIP) > 0 {
		proxy, err := newTrustedProxy(config.TrustedProxiesIP)
		if err != nil {
			logger.Panic("Failed to parse trusted proxies", err)
		}
		trustedProxies = proxy
	}

	return func(ctx context.Context, c *fiber.Ctx) (context.Context, error) {
		// Extract client IP from given header
		if config.TrustedHeader != "" {
			headerIP := c.Get(config.TrustedHeader)

			// validate ip from header
			if ip := net.ParseIP(headerIP); ip != nil {
				return context.WithValue(ctx, clientIPKey{}, headerIP), nil
			}
		}

		// Extract client IP from XFF header
		rawIPs := c.IPs()
		ips := parseIPs(rawIPs)

		// If the request is directly from client, we can use direct remote IP address
		if len(ips) == 0 {
			return context.WithValue(ctx, clientIPKey{}, c.IP()), nil
		}

		// Walk back and find first IP that's not trusted proxy
		if len(trustedProxies) > 0 {
			for i := len(ips) - 1; i >= 0; i-- {
				if !trustedProxies.IsTrusted(ips[i]) {
					return context.WithValue(ctx, clientIPKey{}, ips[i].String()), nil
				}
			}

			// If all IPs are trusted proxies, return first IP in XFF header
			return context.WithValue(ctx, clientIPKey{}, rawIPs[0]), nil
		}

		// Finally, if we can't extract client IP, return forbidden
		if config.EnableRejectMalformedRequest {
			logger.WarnContext(ctx, "IP Spoofing detected, returning 403 Forbidden",
				slog.String("event", "requestcontext/ip_spoofing_detected"),
				slog.String("module", "requestcontext/with_clientip"),
				slog.String("ip", c.IP()),
				slog.Any("ips", rawIPs),
			)
			return nil, requestcontextError{
				status:  fiber.StatusForbidden,
				message: "not allowed to access",
			}
		}

		// Fallback to first IP in XFF header
		return context.WithValue(ctx, clientIPKey{}, rawIPs[0]), nil
	}
}

// GetClientIP get clientIP from context. If not found, return empty string
//
// Warning: Request context should be setup before using this function
func GetClientIP(ctx context.Context) string {
	if ip, ok := ctx.Value(clientIPKey{}).(string); ok {
		return ip
	}
	return ""
}

type trustedProxy []*net.IPNet

// newTrustedProxy create a new trusted proxies instance for preventing IP spoofing (XFF Attacks)
func newTrustedProxy(ranges []string) (trustedProxy, error) {
	nets, err := parseCIDRs(ranges)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return trustedProxy(nets), nil
}

func (t trustedProxy) IsTrusted(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, r := range t {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

func parseCIDRs(ranges []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0, len(ranges))
	for _, r := range ranges {
		_, ipnet, err := net.ParseCIDR(r)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse CIDR for %q", r)
		}
		nets = append(nets, ipnet)
	}
	return nets, nil
}

func parseIPs(ranges []string) []net.IP {
	ip := make([]net.IP, 0, len(ranges))
	for _, r := range ranges {
		ip = append(ip, net.ParseIP(r))
	}
	return ip
}
