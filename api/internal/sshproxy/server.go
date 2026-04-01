package sshproxy

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	glssh "github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	listenAddr      string
	hostKeyPath     string
	containerPrefix string
	docker          *client.Client
	network         string
}

func NewServer(listenAddr, hostKeyPath, containerPrefix string, docker *client.Client, network string) *Server {
	if hostKeyPath == "" {
		hostKeyPath = "storage/ssh_host_key"
	}
	return &Server{
		listenAddr:      listenAddr,
		hostKeyPath:     hostKeyPath,
		containerPrefix: containerPrefix,
		docker:          docker,
		network:         network,
	}
}

func (s *Server) ListenAndServe() error {
	srv := &glssh.Server{
		Addr:    s.listenAddr,
		Handler: s.handleSession,
		SubsystemHandlers: map[string]glssh.SubsystemHandler{
			"sftp": s.handleSession,
		},
		PasswordHandler: func(ctx glssh.Context, password string) bool {
			ctx.SetValue(ctxKeyPassword, password)
			return true
		},
	}

	if err := s.ensureHostKey(srv); err != nil {
		slog.Warn("SSH proxy: persistent host key unavailable, using ephemeral", "error", err)
	}

	slog.Info("SSH proxy listening", "addr", s.listenAddr)
	return srv.ListenAndServe()
}

func (s *Server) ensureHostKey(srv *glssh.Server) error {
	if _, err := os.Stat(s.hostKeyPath); err == nil {
		return srv.SetOption(glssh.HostKeyFile(s.hostKeyPath))
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.hostKeyPath), 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.OpenFile(s.hostKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return fmt.Errorf("encode key: %w", err)
	}
	return srv.SetOption(glssh.HostKeyFile(s.hostKeyPath))
}

type contextKey string

const ctxKeyPassword contextKey = "password"

func parseUser(raw string) (sshUser, sandboxID string) {
	if i := strings.LastIndex(raw, "+"); i > 0 && i < len(raw)-1 {
		return raw[:i], raw[i+1:]
	}
	return "", raw
}

func (s *Server) handleSession(sess glssh.Session) {
	sshUser, sandboxID := parseUser(sess.User())
	if sandboxID == "" {
		_, _ = fmt.Fprintln(sess.Stderr(), "error: no sandbox ID provided")
		_ = sess.Exit(1)
		return
	}

	containerName := s.containerPrefix + sandboxID
	password, _ := sess.Context().Value(ctxKeyPassword).(string)

	upstream, err := s.dialUpstream(sess.Context(), containerName, sshUser, password)
	if err != nil {
		slog.Warn("SSH proxy: dial failed", "container", containerName, "error", err)
		_, _ = fmt.Fprintf(sess.Stderr(), "error: %s\r\n", err)
		_ = sess.Exit(1)
		return
	}
	defer func() { _ = upstream.Close() }()

	upstreamSess, err := upstream.NewSession()
	if err != nil {
		_, _ = fmt.Fprintf(sess.Stderr(), "error: session failed: %s\r\n", err)
		_ = sess.Exit(1)
		return
	}
	defer func() { _ = upstreamSess.Close() }()

	if err := s.setupPTY(sess, upstreamSess); err != nil {
		_, _ = fmt.Fprintf(sess.Stderr(), "error: %s\r\n", err)
		_ = sess.Exit(1)
		return
	}

	stdin, _ := upstreamSess.StdinPipe()
	stdout, _ := upstreamSess.StdoutPipe()
	stderr, _ := upstreamSess.StderrPipe()

	if err := s.startUpstream(sess, upstreamSess); err != nil {
		_, _ = fmt.Fprintf(sess.Stderr(), "error: %s\r\n", err)
		_ = sess.Exit(1)
		return
	}

	_ = sess.Exit(s.proxyIO(sess, upstreamSess, stdin, stdout, stderr))
}

func (s *Server) dialUpstream(ctx context.Context, containerName, sshUser, password string) (*gossh.Client, error) {
	resolveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addr, labelUser, err := s.resolveTarget(resolveCtx, containerName)
	if err != nil {
		return nil, err
	}

	if sshUser == "" {
		sshUser = labelUser
	}
	if sshUser == "" {
		return nil, fmt.Errorf("no SSH user: provide user+sandboxID or set sandbox_ssh_username label")
	}

	var auth []gossh.AuthMethod
	if password != "" {
		auth = append(auth, gossh.Password(password))
	}

	slog.Debug("SSH proxy: connecting", "target", addr, "user", sshUser)
	return gossh.Dial("tcp", addr, &gossh.ClientConfig{
		User: sshUser,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	})
}

func (s *Server) setupPTY(sess glssh.Session, upstream *gossh.Session) error {
	ptyReq, winCh, isPty := sess.Pty()
	if !isPty {
		return nil
	}
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := upstream.RequestPty(ptyReq.Term, ptyReq.Window.Height, ptyReq.Window.Width, modes); err != nil {
		return fmt.Errorf("PTY request failed: %w", err)
	}
	go func() {
		for win := range winCh {
			_ = upstream.WindowChange(win.Height, win.Width)
		}
	}()
	return nil
}

func (s *Server) startUpstream(sess glssh.Session, upstream *gossh.Session) error {
	if sub := sess.Subsystem(); sub != "" {
		return upstream.RequestSubsystem(sub)
	}
	if cmd := sess.RawCommand(); cmd != "" {
		return upstream.Start(cmd)
	}
	return upstream.Shell()
}

func (s *Server) proxyIO(sess glssh.Session, upstream *gossh.Session, stdin io.WriteCloser, stdout, stderr io.Reader) int {
	done := make(chan struct{})

	if stdin != nil {
		go func() {
			_, _ = io.Copy(stdin, sess)
			_ = stdin.Close()
		}()
	}
	if stdout != nil {
		go func() {
			_, _ = io.Copy(sess, stdout)
			close(done)
		}()
	} else {
		close(done)
	}
	if stderr != nil {
		go func() { _, _ = io.Copy(sess.Stderr(), stderr) }()
	}

	exitCode := 0
	if err := upstream.Wait(); err != nil {
		if e, ok := err.(*gossh.ExitError); ok {
			exitCode = e.ExitStatus()
		}
	}

	<-done
	_ = upstream.Close()
	return exitCode
}

func (s *Server) resolveTarget(ctx context.Context, containerName string) (addr, labelUser string, err error) {
	info, err := s.docker.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", "", fmt.Errorf("container %q not found", containerName)
	}

	portStr := info.Config.Labels["sandbox_ssh_port"]
	if portStr == "" {
		return "", "", fmt.Errorf("container %q does not support SSH", containerName)
	}
	labelUser = info.Config.Labels["sandbox_ssh_username"]

	sshNatPort := nat.Port(portStr + "/tcp")
	if bindings, ok := info.NetworkSettings.Ports[sshNatPort]; ok && len(bindings) > 0 && bindings[0].HostPort != "" {
		return net.JoinHostPort("127.0.0.1", bindings[0].HostPort), labelUser, nil
	}

	targetPort, _ := strconv.Atoi(portStr)
	if s.network != "" {
		if ep, ok := info.NetworkSettings.Networks[s.network]; ok && ep.IPAddress != "" {
			return net.JoinHostPort(ep.IPAddress, strconv.Itoa(targetPort)), labelUser, nil
		}
	}
	for _, ep := range info.NetworkSettings.Networks {
		if ep.IPAddress != "" {
			return net.JoinHostPort(ep.IPAddress, strconv.Itoa(targetPort)), labelUser, nil
		}
	}

	return "", "", fmt.Errorf("container %q: no reachable SSH endpoint", containerName)
}
