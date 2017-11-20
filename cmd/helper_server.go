package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

func refreshRules(c *proxyConfig, m *rule.CachedMatcher, fails int) {
	duration, _ := time.ParseDuration(c.refreshDelay)
	if duration == 0 {
		duration = time.Second * 30
	}

	if err := m.Refresh(); err != nil {
		logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh rules")
		if fails > 15 {
			logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d\n", fails)
		}

		time.Sleep(time.Second * time.Duration(fails+1))
		refreshRules(c, m, fails+1)
		return
	}

	time.Sleep(duration)

	refreshRules(c, m, 0)
}

func refreshKeys(k rsakey.Manager, fails int) {
	duration := time.Minute * 5

	if err := k.Refresh(); err != nil {
		logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh RSA keys for JWK signing")
		if fails > 15 {
			logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d\n", fails)
		}

		time.Sleep(time.Second * time.Duration(fails+1))
		refreshKeys(k, fails+1)
		return
	}

	time.Sleep(duration)

	refreshKeys(k, 0)
}

func parseCorsOptions(prefix string) cors.Options {
	if prefix != "" {
		prefix = prefix + "_"
	}

	allowCredentials, _ := strconv.ParseBool(viper.GetString(prefix + "CORS_ALLOWED_CREDENTIALS"))
	debug, _ := strconv.ParseBool(viper.GetString(prefix + "CORS_DEBUG"))
	maxAge, _ := strconv.Atoi(viper.GetString(prefix + "CORS_MAX_AGE"))
	return cors.Options{
		AllowedOrigins:   strings.Split(viper.GetString(prefix+"CORS_ALLOWED_ORIGINS"), ","),
		AllowedMethods:   strings.Split(viper.GetString(prefix+"CORS_ALLOWED_METHODS"), ","),
		AllowedHeaders:   strings.Split(viper.GetString(prefix+"CORS_ALLOWED_HEADERS"), ","),
		ExposedHeaders:   strings.Split(viper.GetString(prefix+"CORS_EXPOSED_HEADERS"), ","),
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
		Debug:            debug,
	}
}