/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/inhies/go-bytesize"
	"golang.org/x/crypto/ssh"
)

const (
	// https://tools.ietf.org/html/rfc952
	HostnameRegexStringRFC952 = `^[a-zA-Z][a-zA-Z0-9\-\.]+[a-zA-Z0-9]$`
	// accepts hostname starting with a digit https://tools.ietf.org/html/rfc1123
	HostnameRegexStringRFC1123 = `^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*?$`

	PortErrorExpression   = "Invalid port list expression"
	PortErrorInvalid      = "One of the specified ports is"
	PortErrorRangeSize    = "Port range has more than 2 elements"
	PortErrorRangeInvalid = "One of the specified port range has last port > first port"

	MinPort = 1
	MaxPort = 65535
)

func SetDefaultStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func SetFieldStr(field *string, value string) {
	if value != "" {
		*field = value
	}
}

func XmlMarshal(b interface{}) (string, error) {
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("  ", "    ")
	err := enc.Encode(b)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func XmlUnmarshal(input string, v any) error {
	return xml.Unmarshal([]byte(input), v)
}

func HumanByteSize(n uint64) string {
	b := bytesize.New(float64(n))
	return b.String()
}

func byteCountIEC(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

var (
	emailVerifier = emailverifier.NewVerifier()
	domainRegex   = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	hostnameRegex = regexp.MustCompile(HostnameRegexStringRFC1123)
	usernameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
)

func VerifyEmail(email string) error {
	syntax := emailVerifier.ParseAddress(email)
	if !syntax.Valid {
		return fmt.Errorf("email address %s syntax is invalid", email)
	}

	return nil
}

func VerifyDomain(name string) bool {
	return domainRegex.MatchString(name)
}

func VerifyHostname(name string) bool {
	return hostnameRegex.MatchString(name)
}

// VerifyUnixUsername tells whether name is a portable Unix user name, safe to
// be rendered into cloud-init templates.
func VerifyUnixUsername(name string) bool {
	return usernameRegex.MatchString(name)
}

// VerifySSHPublicKey ensures key is a single, valid authorized_keys entry, safe
// to be rendered into cloud-init templates.
func VerifySSHPublicKey(key string) error {
	if strings.ContainsAny(key, "\r\n") {
		return fmt.Errorf("SSH public key must fit on a single line")
	}
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return fmt.Errorf("invalid SSH public key: %w", err)
	}
	return nil
}

// verifyBootstrapSettings validates the service user settings that get
// rendered into instances cloud-init user-data.
func verifyBootstrapSettings(user, pubkey string) error {
	if user != "" && !VerifyUnixUsername(user) {
		return fmt.Errorf("invalid bootstrap user name: %s", user)
	}
	if pubkey != "" {
		return VerifySSHPublicKey(pubkey)
	}
	return nil
}

func HasChildRef(children *[]string, childId string) bool {
	return slices.Contains(*children, childId)
}

func FindChildByID[T any](children *[]string, childId, collection, msg string) (*T, error) {
	var res T
	if !HasChildRef(children, childId) {
		return &res, errors.New(msg)
	}

	return FindResourceByID[T](collection, childId)
}

func AddChildRef(children *[]string, childId string) {
	if !HasChildRef(children, childId) {
		*children = append(*children, childId)
	}
}

func RemoveChildRef(children *[]string, childId string) {
	if idx := slices.Index(*children, childId); idx != -1 {
		*children = slices.Delete(*children, idx, idx+1)
	}
}

func HasChildRefs(children ...[]string) bool {
	for _, c := range children {
		if len(c) > 0 {
			return true
		}
	}
	return false
}

func bytesToGB(val int64) int64 {
	return val / 1024 / 1024 / 1024
}

func FindNextBitIp(ipnet net.IPNet, bit int) (net.IP, error) {
	if bit < 0 {
		return ipnet.IP, fmt.Errorf("IP incrementation is outside of subnet %s", ipnet.String())
	}

	ip4 := ipnet.IP.To4()
	if ip4 != nil {
		targetVal := binary.BigEndian.Uint32(ip4) + uint32(bit) // #nosec G115 -- bit is verified non-negative and fits in uint32
		target := make(net.IP, 4)
		binary.BigEndian.PutUint32(target, targetVal)
		if ipnet.Contains(target) {
			return target, nil
		}
		return target, fmt.Errorf("IP incrementation is outside of subnet %s", ipnet.String())
	}

	netcidr_first_ip := ipnet.IP
	for i := 0; i < bit; i++ {
		netcidr_first_ip = cidr.Inc(netcidr_first_ip)
	}
	if ipnet.Contains(netcidr_first_ip) {
		return netcidr_first_ip, nil
	}
	return netcidr_first_ip, fmt.Errorf("IP incrementation is outside of subnet %s", ipnet.String())
}

func isValidPort(expr, p string) (int, error) {
	port, err := strconv.Atoi(p)
	if err != nil {
		return -1, err
	}
	if port < MinPort {
		return port, fmt.Errorf("%s: '%s'. %s < %d", PortErrorExpression, expr, PortErrorInvalid, MinPort)
	}
	if port > MaxPort {
		return port, fmt.Errorf("%s: '%s'. %s > %d", PortErrorExpression, expr, PortErrorInvalid, MaxPort)
	}
	return port, nil
}

func IsValidPortListExpression(expr string) error {
	items := strings.Split(expr, ",")
	for _, i := range items {
		rg := strings.Split(i, "-")
		if len(rg) > 2 {
			return fmt.Errorf("%s: '%s'. %s", PortErrorExpression, expr, PortErrorRangeSize)
		}
		first, err := isValidPort(expr, rg[0])
		if err != nil {
			return err
		}
		if len(rg) == 2 {
			last, err := isValidPort(expr, rg[1])
			if err != nil {
				return err
			}
			if last < first {
				return fmt.Errorf("%s: '%s'. %s", PortErrorExpression, expr, PortErrorRangeInvalid)
			}
		}
	}

	return nil
}

// ptr returns a pointer to the given value, used for optional API model fields.
func ptr[T any](v T) *T {
	return &v
}

// valueOr returns the optional API model field value, or the given default if
// unspecified.
func valueOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// toInt32 converts a 64-bit integer to an API model 32-bit one, clamping out
// of range values.
func toInt32(v int64) int32 {
	return int32(max(min(v, math.MaxInt32), math.MinInt32)) // #nosec G115 -- clamped
}
