package main

import (
	"fmt"
	"github.com/containrrr/shoutrrr"
	stypes "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"shoutrrr-api/middleware"
	"shoutrrr-api/types"
	"time"
)

type Config struct {
	Port     int `yaml:"port"`
	LogLevel int `yaml:"logLevel"`
	Config   []struct {
		ID    string   `yaml:"id"`
		Shout []string `yaml:"shout"`
	} `yaml:"config"`
}

var config = Config{}

func ParseError(c *fiber.Ctx) error {
	log.Warnf("failed to parse request")
	return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
		"error":   true,
		"message": "failed to parse request",
	})
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)

	dat, err := os.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		panic(err)
	}

	log.SetLevel(log.AllLevels[config.LogLevel])

	for _, entry := range config.Config {
		for _, url := range entry.Shout {
			_, err := shoutrrr.CreateSender(url)
			if err != nil {
				log.WithError(err).WithField("url", url).Panic("failed to parse shoutrrr url")
			}
		}
	}

	app := fiber.New()
	app.Use(middleware.Logger())

	app.Post("/:id/:type?", func(c *fiber.Ctx) error {
		return c.JSON(time.Now().Format(time.RFC3339))
	})

	app.Post("/:id/:type?", func(c *fiber.Ctx) error {
		var title = ""
		var msg = ""

		var shout []string
		for _, entry := range config.Config {
			if entry.ID == c.Params("id") {
				shout = entry.Shout
				break
			}
		}
		if len(shout) == 0 {
			log.WithField("id", c.Params("id")).Warnf("unknown id request")
			return c.Status(fiber.StatusForbidden).JSON(map[string]interface{}{
				"error":   true,
				"message": "unknown id",
			})
		}

		log.WithFields(log.Fields{
			"uri":          c.Request().URI().String(),
			"body":         string(c.Body()),
			"content-type": c.Get(fiber.HeaderContentType),
		}).Tracef("new request")

		switch c.Params("type") {
		case "slack":
			req := new(types.Payload)
			if err := c.BodyParser(req); err != nil {
				log.Warnf("unknown slack body")
				return ParseError(c)
			}
			log.WithField("req", req).Debug("new slack request")
			msg = req.Text
			for _, attachment := range req.Attachments {
				if len(msg) == 0 {
					msg = attachment.Title
				} else {
					msg = "\r\n" + attachment.Title
				}

				for _, embedField := range attachment.Fields {
					if len(msg) == 0 {
						msg = embedField.Title + ": " + embedField.Value
					} else {
						msg = "\r\n" + embedField.Title + ": " + embedField.Value
					}
				}

				if len(msg) == 0 {
					msg = attachment.Footer
				} else {
					msg = "\r\n" + attachment.Footer
				}
			}
			break
		case "discord":
			req := new(types.Message)
			if err := c.BodyParser(req); err != nil {
				log.Warnf("unknown discord body")
				return ParseError(c)
			}
			log.WithField("req", req).Debug("new discord request")
			msg = req.Content
			for _, embed := range req.Embeds {
				if len(msg) == 0 {
					msg = embed.Title
				} else {
					msg = "\r\n" + embed.Title
				}

				if len(msg) == 0 {
					msg = embed.Description
				} else {
					msg = "\r\n" + embed.Description
				}

				for _, embedField := range embed.Fields {
					if len(msg) == 0 {
						msg = embedField.Name + ": " + embedField.Value
					} else {
						msg = "\r\n" + embedField.Name + ": " + embedField.Value
					}

				}

				if len(msg) == 0 {
					msg = embed.Footer.Text
				} else {
					msg = "\r\n" + embed.Footer.Text
				}
			}
			break
		default:
			req := new(types.Simple)
			if err := c.BodyParser(req); err != nil {
				log.Warnf("unknown default body defaulting to body")
				msg = string(c.Body())
			} else {
				log.WithField("req", req).Debug("new default request")
				msg = req.Message
				title = req.Title
			}

			if len(title) == 0 {
				title = c.Query("title", c.Query("_title", "Unknown"))
			}
		}

		if len(shout) > 0 {

			sender, err := shoutrrr.CreateSender(shout...)
			if err != nil {
				log.WithError(err).Warnf("failed to create shoutrrr sender")
				return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
					"error":   true,
					"message": err.Error(),
				})
			}
			errs := sender.Send(msg, &stypes.Params{
				stypes.TitleKey: title,
			})

			var errorString []string
			for _, theError := range errs {
				if theError == nil {
					continue
				}
				errorString = append(errorString, theError.Error())
			}
			if len(errorString) > 0 {
				log.WithField("errors", errorString).Warnf("failed to send shoutrrr message")
				return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
					"error":   true,
					"message": errorString,
				})
			}
		} else {
			log.WithField("id", c.Params("id")).Warnf("empty shout")
		}

		log.WithField("id", c.Params("id")).Debug("new shout send")
		return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"error":   false,
			"message": "ok",
		})
	})

	app.Listen(fmt.Sprintf(":%d", config.Port))
}
