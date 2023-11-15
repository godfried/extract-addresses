package email

import (
	"fmt"
	"io"
	"net/mail"
	"strings"

	"github.com/mnako/letters"
	"golang.org/x/net/html"
)

type Address struct {
	Email   *mail.Address
	Context AddressContext
}

type AddressContext int

const (
	AddressContextNone AddressContext = iota
	AddressContextFrom
	AddressContextTo
	AddressContextForwardedFrom
	AddressContextForwardedTo
	AddressContextMax
)

func (c *AddressContext) Set(val string) error {
	*c = GetContext(val, false)
	return nil
}

func (c AddressContext) String() string {
	switch c {
	case AddressContextNone:
		return "None"
	case AddressContextFrom:
		return "From"
	case AddressContextTo:
		return "To"
	case AddressContextForwardedFrom:
		return "ForwardedFrom"
	case AddressContextForwardedTo:
		return "ForwardedTo"
	default:
		return "Unknown"
	}
}

func GetContext(s string, forwarded bool) AddressContext {
	m := strings.TrimSuffix(strings.ToLower(s), ":")
	switch m {
	case "from":
		if forwarded {
			return AddressContextForwardedFrom
		}
		return AddressContextFrom
	case "to":
		if forwarded {
			return AddressContextForwardedTo
		}
		return AddressContextTo
	case "forwardedfrom":
		return AddressContextForwardedFrom
	case "forwardedto":
		return AddressContextForwardedTo
	}
	return AddressContextNone
}

func Parse(r io.Reader) (addresses [AddressContextMax][]Address, err error) {
	email, err := letters.ParseEmail(r)
	if err != nil {
		return addresses, fmt.Errorf("failed to parse email: %e", err)
	}
	parsed, err := ParseHTML(email.HTML)
	return parsed, err
}

func ParseHTML(htmlContent string) (addresses [AddressContextMax][]Address, err error) {
	parsedHTML, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return addresses, err
	}
	var f func(*html.Node) error
	f = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "font" {
			if n.FirstChild == nil || n.FirstChild.FirstChild == nil || n.FirstChild.NextSibling == nil {
				return nil
			}
			fromTo := n.FirstChild.FirstChild.Data
			context := GetContext(fromTo, true)
			emailData := strings.TrimSpace(n.FirstChild.NextSibling.Data)
			if context == AddressContextNone {
				return nil
			}
			email, err := mail.ParseAddress(emailData)
			if err != nil {
				emailData = strings.Split(emailData, " ")[0]
				email, err = mail.ParseAddress(emailData)
				if err != nil {
					return err
				}
			}
			addresses[context] = append(addresses[context], Address{
				Email:   email,
				Context: context,
			})
			return nil
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := f(c); err != nil {
				return err
			}
		}
		return nil
	}
	err = f(parsedHTML)
	return addresses, err
}
