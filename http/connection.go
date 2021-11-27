// Package http provides low level abstractions for sending/receiving raw http messages
package http

import (
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// DestinationFromString create a Destination from String
func DestinationFromString(urlString string) *Destination {
	u, _ := url.Parse(urlString)
	host, port, _ := net.SplitHostPort(u.Host)
	p, _ := strconv.Atoi(port)

	d := &Destination{
		Port:     p,
		DestAddr: host,
		Protocol: u.Scheme,
	}

	return d
}

// StartTrackingTime initializes timer
func (c *Connection) StartTrackingTime() {
	c.duration.StartTracking()
}

// StopTrackingTime stops timer
func (c *Connection) StopTrackingTime() {
	c.duration.StopTracking()
}

// GetRoundTripTime will return the time since the request started and the response was parsed
func (c *Connection) GetRoundTripTime() *RoundTripTime {
	return c.duration
}

func (c *Connection) send(data []byte) (int, error) {
	var err error
	var sent int

	log.Trace().Msg("ftw/http: sending data")
	// Store times for searching in logs, if necessary

	if c.connection != nil {
		sent, err = c.connection.Write(data)
	} else {
		err = errors.New("ftw/http/send: not connected to server")
	}

	return sent, err

}

func (c *Connection) receive() ([]byte, error) {
	log.Trace().Msg("ftw/http: receiving data")
	var err error
	var buf []byte

	// Set a deadline for reading. Read operation will fail if no data
	// is received after deadline.
	timeoutDuration := 1000 * time.Millisecond

	// We assume the response body can be handled in memory without problems
	// That's why we use io.ReadAll
	if err = c.connection.SetReadDeadline(time.Now().Add(timeoutDuration)); err == nil {
		buf, err = io.ReadAll(c.connection)
	}

	if neterr, ok := err.(net.Error); ok && !neterr.Timeout() {
		log.Error().Msgf("ftw/http: %s\n", err.Error())
	} else {
		err = nil
	}
	log.Trace().Msgf("ftw/http: received data - %q", buf)

	return buf, err
}
