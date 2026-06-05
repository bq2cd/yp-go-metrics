package httpheaders

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// HeaderKeyXRealIP represents the key for 'X-Real-IP' HTTP header
const HeaderKeyXRealIP = "X-Real-IP"

const xRealIPHashParameterPrefix = "hash="

// XRealIP represents value of `X-Real-IP` header, which contains
// IP address and optional HMAC signature for IP address value.
type XRealIP struct {
	IP   net.IP
	Hash []byte
}

// GetXRealIP parses `X-Real-IP` header in the form of `IP;hash=HASH` (where `;hash=HASH` is optional)
// and returns populated [XRealIP] struct.
func GetXRealIP(header http.Header) XRealIP {
	return GetXRealIPFromBytes([]byte(header.Get(HeaderKeyXRealIP)))
}

// GetXRealIPFromBytes parses provided data bytes in the form of `IP;hash=HASH` (where `;hash=HASH` is optional)
// and returns populated [XRealIP] struct.
func GetXRealIPFromBytes(data []byte) XRealIP {
	var p xRealIPParser

	return p.Parse(data)
}

// String converts [XRealIP] to a header value in the format of `IP;hash=Hash`
// (where `;hash=Hash` is only appended when [XRealIP.Hash] is not empty).
func (h XRealIP) String() string {
	if len(h.IP) == 0 {
		return ""
	}

	sb := new(strings.Builder)

	fmt.Fprint(sb, h.IP.String())
	if len(h.Hash) > 0 {
		fmt.Fprint(sb, ";", xRealIPHashParameterPrefix, hex.EncodeToString(h.Hash))
	}

	return sb.String()
}

// Empty returns `true` if [XRealIP] has no IP address and no hash.
func (h XRealIP) Empty() bool {
	return len(h.IP) == 0 && len(h.Hash) == 0
}

// Equal returns `true` if two [XRealIP] structs have the same IP address and the same hash (even if empty).
func (h XRealIP) Equal(other XRealIP) bool {
	return h.IP.Equal(other.IP) && bytes.Equal(h.Hash, other.Hash)
}

// Matches returns `true` if current [XRealIP] matches the value decoded from provided HTTP header.
func (h XRealIP) Matches(header http.Header) bool {
	other := GetXRealIP(header)

	return h.Equal(other)
}

// Apply adds [XRealIP] to HTTP header, encoded as string; all previous values of the header are
// overwritten. If [XRealIP] is empty, HTTP header is removed instead.
func (h XRealIP) Apply(header http.Header) XRealIP {
	if h.Empty() {
		header.Del(HeaderKeyXRealIP)
	} else {
		header.Set(HeaderKeyXRealIP, h.String())
	}

	return h
}

type xRealIPParser struct {
	buf            []byte
	paramPositions []int
	out            XRealIP
}

// Parse parses provided data bytes in the form of `IP;hash=HASH` (where `;hash=HASH` is optional)
// and returns populated [XRealIP] struct.
func (p *xRealIPParser) Parse(data []byte) XRealIP {
	p.init(len(data))

	p.removeSpacesAndRecordParamPositions(data)
	p.extractIP()
	p.extractHash()

	return p.out
}

func (p *xRealIPParser) init(bufSize int) {
	p.buf = make([]byte, 0, bufSize)
	p.paramPositions = make([]int, 0, 2) // at least 1 parameter is expected
	p.out = XRealIP{}
}

func (p *xRealIPParser) removeSpacesAndRecordParamPositions(data []byte) {
	// remove all spaces before extracting
	for _, char := range data {
		switch char {
		case ' ':
			// drop spaces
		case ';':
			p.paramPositions = append(p.paramPositions, len(p.buf))
		default:
			p.buf = append(p.buf, char)
		}
	}
}

func (p *xRealIPParser) extractIP() {
	startParamPos := len(p.buf)
	if len(p.paramPositions) > 0 {
		startParamPos = p.paramPositions[0]
	}

	p.out.IP = net.ParseIP(string(p.buf[:startParamPos]))
}

func (p *xRealIPParser) extractHash() {
	if len(p.paramPositions) == 0 {
		return
	}

	// extract all params into map; first seen wins
	params := make(map[string][]byte, len(p.paramPositions))

	for i := 1; i <= len(p.paramPositions); i++ {
		start := p.paramPositions[i-1]
		next := len(p.buf)
		if i < len(p.paramPositions) {
			next = p.paramPositions[i]
		}

		sep := bytes.Index(p.buf[start:next], []byte(`=`))
		if sep <= 0 { // params without key are invalid anyway
			continue
		}

		// include `=` into param key
		key := string(p.buf[start : start+sep+1])
		if _, ok := params[key]; !ok { // first seen wins
			params[key] = p.buf[start+sep+1 : next]
		}
	}

	if hash, ok := params[xRealIPHashParameterPrefix]; ok {
		// it's okay if hash is not valid hex-encoded string;
		// this will be caught by hash validation anyway;
		p.out.Hash, _ = hex.AppendDecode(p.out.Hash, hash)
	}
}
