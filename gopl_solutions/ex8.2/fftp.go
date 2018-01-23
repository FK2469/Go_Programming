/* 
ex8.2 is a minimal ftp server as per section 5.1 of RFC 959.
5.1.  MINIMUM IMPLEMENTATION

In order to make FTP workable without needless error messages, the
following minimum implementation is required for all servers:

   TYPE - ASCII Non-print
   MODE - Stream
   STRUCTURE - File, Record
   COMMANDS - USER, QUIT, PORT,
			  TYPE, MODE, STRU,
				for the default values
			  RETR, STOR,
			  NOOP.

The default values for transfer parameters are:

   TYPE - ASCII Non-print
   MODE - Stream
   STRU - File

All hosts must accept the above as the standard defaults.

USER NAME (USER)
	The argument field is a Telnet string identifying the user.
	The user identification is that which is required by the
	server for access to its file system.  This command will
	normally be the first command transmitted by the user after
	the control connections are made (some servers may require
	this).  Additional identification information in the form of
	a password and/or an account command may also be required by
	some servers.  Servers may allow a new USER command to be
	entered at any point in order to change the access control
	and/or accounting information.  This has the effect of
	flushing any user, password, and account information already
	supplied and beginning the login sequence again.  All
	transfer parameters are unchanged and any file transfer in
	progress is completed under the old access control
	parameters.

LOGOUT (QUIT)
	This command terminates a USER and if file transfer is not
	in progress, the server closes the control connection.  If
	file transfer is in progress, the connection will remain
	open for result response and the server will then close it.
	If the user-process is transferring files for several USERs
	but does not wish to close and then reopen connections for
	each, then the REIN command should be used instead of QUIT.

	An unexpected close on the control connection will cause the
	server to take the effective action of an abort (ABOR) and a
	logout (QUIT)

DATA PORT (PORT)
	The argument is a HOST-PORT specification for the data port
	to be used in data connection.  There are defaults for both
	the user and server data ports, and under normal
	circumstances this command and its reply are not needed.  If
	this command is used, the argument is the concatenation of a
	32-bit internet host address and a 16-bit TCP port address.
	This address information is broken into 8-bit fields and the
	value of each field is transmitted as a decimal number (in
	character string representation).  The fields are separated
	by commas.  A port command would be:

		PORT h1,h2,h3,h4,p1,p2

	where h1 is the high order 8 bits of the internet host
	address.

PASSIVE (PASV)
	This command requests the server-DTP to "listen" on a data
	port (which is not its default data port) and to wait for a
	connection rather than initiate one upon receipt of a
	transfer command.  The response to this command includes the
	host and port address this server is listening on.

REPRESENTATION TYPE (TYPE)
	The argument specifies the representation type as described
	in the Section on Data Representation and Storage.  Several
	types take a second parameter.  The first parameter is
	denoted by a single Telnet character, as is the second
	Format parameter for ASCII and EBCDIC; the second parameter
	for local byte is a decimal integer to indicate Bytesize.
	The parameters are separated by a <SP> (Space, ASCII code
	32).

	The following codes are assigned for type:

					\    /
		A - ASCII |    | N - Non-print
					|-><-| T - Telnet format effectors
		E - EBCDIC|    | C - Carriage Control (ASA)
					/    \
		I - Image

		L <byte size> - Local byte Byte size

		The default representation type is ASCII Non-print.  If the
		Format parameter is changed, and later just the first
		argument is changed, Format then returns to the Non-print
		default.

FILE STRUCTURE (STRU)
	The argument is a single Telnet character code specifying
	file structure described in the Section on Data
	Representation and Storage.

	The following codes are assigned for structure:

		F - File (no record structure)
		R - Record structure
		P - Page structure

	The default structure is File.

TRANSFER MODE (MODE)
	The argument is a single Telnet character code specifying
	the data transfer modes described in the Section on
	Transmission Modes.

	The following codes are assigned for transfer modes:

		S - Stream
		B - Block
		C - Compressed

	The default transfer mode is Stream.

RETRIEVE (RETR)
	This command causes the server-DTP to transfer a copy of the
	file, specified in the pathname, to the server- or user-DTP
	at the other end of the data connection.  The status and
	contents of the file at the server site shall be unaffected.

STORE (STOR)
	This command causes the server-DTP to accept the data
	transferred via the data connection and to store the data as
	a file at the server site.  If the file specified in the
	pathname exists at the server site, then its contents shall
	be replaced by the data being transferred.  A new file is
	created at the server site if the file specified in the
	pathname does not already exist.

LIST (LIST)
	This command causes a list to be sent from the server to the
	passive DTP.  If the pathname specifies a directory or other
	group of files, the server should transfer a list of files
	in the specified directory.  If the pathname specifies a
	file then the server should send current information on the
	file.  A null argument implies the user's current working or
	default directory.  The data transfer is over the data
	connection in type ASCII or type EBCDIC.  (The user must
	ensure that the TYPE is appropriately ASCII or EBCDIC).
	Since the information on a file may vary widely from system
	to system, this information may be hard to use automatically
	in a program, but may be quite useful to a human user.

NOOP (NOOP)
	This command does not affect any parameters or previously
	entered commands. It specifies no action other than that the
	server send an OK reply.	

There's a contradiction above for the STRU command. Only File structure (the default) is supported.

Additionally, passive transfers are supported.

Only IPv4 is supported.

DJB's recommendations at http://cr.yp.to/ftp.html have mostly been implemented when noticed and applicable.

*/

package main

import(
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type conn struct{
	rw				net.Conn // "Protocol Interpreter" connection
	dataHostPort	string
	prevCmd			string
	pasvListener	net.Listener
	cmdErr			error
	binary			bool
}

func NewConn(cmdConn net.Conn) *conn{
	return &conn{rw: cmdConn}
}

// hostPortToFTP returns a comma-separated, FTP-style address suitable for replying to the PASV command.
func hostPortToFTP(hostport string)(addr string, err error){
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil{
		return "", err
	}
	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil{
		return "", err
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil{
		return "", err
	}
	ip := ipAddr.IP.To4()
	s := fmt.Sprintf("%d,%d,%d,%d,%d,%d", ip[0], ip[1], ip[2], ip[3], port/256, port%256)
	return s, nil
}

func hostPortFromFTP(address string)(string,error){
	var a, b, c, d byte
	var p1, p2 int
	_, err := fmt.Sscanf(address, "%d,%d,%d,%d,%d,%d", &a, &b, &c, &d, &p1, &p2)
	if err != nil{
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d.%d.%d", a, b, c, d, 256 * p1 + p2), nil
}

func (c *conn) dataConn() (conn io.ReadWriteCloser, err error){
	switch c.prevCmd{
	case "PORT":
		conn, err = net.Dial("tcp", c.dataHostPort)
		if err != nil{
			return nil, err
		}
	case "PASV":
		conn, err = c.pasvListener.Accept()
		if err != nil{
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Previous command not PASV or Port")
	}
	return conn, nil
}

type logPairs map[string]interface{}

func (c *conn) log(pairs logPairs){
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "addr=%s", c.rw.RemoteAddr().String())
	for k,v := range pairs{
		fmt.Fprintf(b, " %s=%s", k, v)
	}
	log.Print(b.String())
}

func (c *conn) writeln(s... interface{}){
	if c.cmdErr != nil{
		return
	}
	s = append(s, "\r\n")
	_, c.cmdErr = fmt.Fprint(c.rw, s...)
}

func (c *conn) lineEnding() string{
	if c.binary{
		return "\n"
	}else{
		return "\r\n"
	}
}

func (c *conn) CmdErr() error{
	return c.cmdErr
}

func (c *conn) Close() error{
	err := c.rw.Close()
	if err != nil{
		c.log(logPairs{"err": fmt.Errorf("Closing command connection: %s", err)})
	}
	return err
}

func (c *conn) list(args []string){
	var filename string
	switch len(args){
	case 0:
		filename = "."
	case 1:
		filename = args[0]
	default:
		c.writeln("501 Too many arguments.")
		return
	}
	file, err := os.Open(filename)
	if err != nil{
		c.writeln("550 File not found.")
		return
	}
	c.writeln("150 Here comes the directory listing.")
	w, err := c.dataConn()
	if err != nil{
		c.writeln("425 Can't open data connection.")
		return
	}
	defer w.Close()
	stat, err := file.Stat()
	if err != nil{
		c.log(logPairs{"cmd":"LIST","err":err})
		c.writeln("450 Requested file action not taken. File unavailable.")
	}
	if stat.IsDir(){
		filenames, err := file.Readdirnames(0)
		if err != nil{
			c.writeln("550 Can't read directory.")
			return
		}
		for _,f := range filenames{
			_, err = fmt.Fprint(w, f, c.lineEnding())
			if err != nil{
				c.log(logPairs{"cmd":"LIST","err":err})
				c.writeln("426 Connection closed: transfer aborted.")
				return
			}
		}
	}else{
		_, err = fmt.Fprint(w, filename, c.lineEnding())
		if err != nil{
			c.log(logPairs{"cmd":"LIST","err":err})
			c.writeln("426 Connction closed: transfer aborted.")
			return
		}
	}
	c.writeln("226 Closing data connection. List successful.")
}

func (c *conn) stru(args []string){
	if len(args) != 1{
		c.writeln("501 Usage: STRU F")
		return
	}
	if args[0] != "F"{
		c.writeln("504 Only file structure is supported.")
		return
	}
	c.writeln("200 STRU set")
}

func (c *conn) retr(args []string){
	if len(args) != 1{
		c.writeln("501 Usage: RETR filename.")
		return
	}
	filename := args[0]
	file, err := os.Open(filename)
	if err != nil{
		c.log(logPairs{"cmd":"RETR", "err":err})
		c.writeln("550 file not found.")
		return
	}
	c.writeln("150 File OK. Sending.")
	conn, err := c.dataConn()
	if err != nil{
		c.writeln("425 Can't open data connection.")
		return 
	}
	defer conn.Close()
	if c.binary{
		_, err := io.Copy(conn, file)
		if err != nil{
			c.log(logPairs{"cmd":"RETR","err":err})
			c.writeln("450 File unavailable.")
			return
		}
	} else{
		// Convert line endings LF -> CRLF.
		r := bufio.NewReader(file)
		w := bufio.NewWriter(conn)
		for {
			line, isPrefix, err := r.ReadLine()
			if err != nil{
				if err != io.EOF {
					break
				}
				c.log(logPairs{"cmd":"RETR", "err":err})
				c.writeln("450 File unavailable.")
				return
			}
			w.Write(line)
			if !isPrefix {
				w.Write([]byte("\r\n"))
			}
		}
		w.Flush()
	}
	c.writeln("226 Transfer complete.")
}

func (c *conn) port(args []string){
	if len(args) != 1{
		c.writeln("501 Usage: PORT a,b,c,d,p1,p2.")
		return
	}
	var err error
	c.dataHostPort, err = hostPortFromFTP(args[0])
	if err != nil{
		c.log(logPairs{"cmd":"PORT", "err":err})
		c.writeln("501 Can't parse address.")
		return
	}
	c.writeln("200 PORT command successful.")
}

func (c *conn) type_(args []string){
	if len(args) < 1 || len(args) > 2{
		c.writeln("501 Usage: TYPE takes 1 or 2 arguments.")
		return
	}
	switch strings.ToUpper(strings.Join(args," ")){
	case "A", "A N":
		c.binary = false
	case "I", "L 8":
		c.binary = true
	default:
		c.writeln("504 Unsupported type. Supported types: A, A N, I, L 8.")
		return
	}
	c.writeln("200 TYPE set.")
}

func (c *conn) stor(args []string) {
	if len(args) != 1 {
		c.writeln("501 Usage: STOR filename")
		return
	}
	filename := args[0]
	file, err := os.Create(filename)
	if err != nil {
		c.log(logPairs{"cmd": "STOR", "err": err})
		c.writeln("550 File can't be created.")
		return
	}
	c.writeln("150 Ok to send data.")
	conn, err := c.dataConn()
	if err != nil {
		c.writeln("425 Can't open data connection")
		return
	}
	defer conn.Close()
	_, err = io.Copy(file, conn)
	if err != nil {
		c.log(logPairs{"cmd": "RETR", "err": err})
		c.writeln("450 File unavailable.")
		return
	}
	c.writeln("226 Transfer complete.")
}

func (c *conn) pasv(args []string) {
	if len(args) > 0 {
		c.writeln("501 Too many arguments.")
		return
	}
	var firstError error
	storeFirstError := func(err error) {
		if firstError == nil {
			firstError = err
		}
	}
	var err error
	c.pasvListener, err = net.Listen("tcp4", "")
	storeFirstError(err)
	_, port, err := net.SplitHostPort(c.pasvListener.Addr().String())
	storeFirstError(err)
	ip, _, err := net.SplitHostPort(c.rw.LocalAddr().String())
	storeFirstError(err)
	addr, err := hostPortToFTP(fmt.Sprintf("%s:%s", ip, port))
	storeFirstError(err)
	if firstError != nil {
		c.pasvListener.Close()
		c.pasvListener = nil
		c.log(logPairs{"cmd": "PASV", "err": err})
		c.writeln("451 Requested action aborted. Local error in processing.")
		return
	}
	// DJB recommends putting an extra character before the address.
	c.writeln(fmt.Sprintf("227 =%s", addr))
}

func (c *conn) run(){
	c.writeln("220 Ready.")
	s := bufio.NewScanner(c.rw)
	var cmd string
	var args []string
	for s.Scan(){
		if c.CmdErr() != nil{
			c.log(logPairs{"err": fmt.Errorf("command connection: %s", c.CmdErr())})
			return
		}
		fields := strings.Fields(s.Text())
		if len(fields) == 0 {
			continue
		}
		cmd = strings.ToUpper(fields[0])
		args = nil
		if len(fields) > 1{
			args = fields[1:]
		}
		switch cmd{
		case "LIST":
			c.list(args)
		case "NOOP":
			c.writeln("200 Ready.")
		case "PASV":
			c.pasv(args)
		case "PORT":
			c.port(args)
		case "QUIT":
			c.writeln("221 Goodbye.")
			return
		case "RETR":
			c.retr(args)
		case "STOR":
			c.stor(args)
		case "STRU":
			c.stru(args)
		case "SYST":
			c.writeln("215 UNIX Type: L8")
		case "TYPE":
			c.type_(args)
		case "USER":
			c.writeln("230 Login successful.")
		default:
			c.writeln(fmt.Sprintf("502 Command %q not implemented.", cmd))
		}
		// Cleanup PASV listeners if they go unused.
		if cmd != "PASV" && c.pasvListener != nil{
			c.pasvListener.Close()
			c.pasvListener = nil
		}
		c.prevCmd = cmd
	}
	if s.Err() != nil{
		c.log(logPairs{"err":fmt.Errorf("scanning commands: %s", s.Err())})
	}
}

func main(){
	var port int
	flag.IntVar(&port, "port", 8000, "listen port")

	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil{
		log.Fatal("Opening main listener:", err)
	}
	for{
		c, err := ln.Accept()
		if err != nil{
			log.Print("Accepting new connection:", err)
		}
		go NewConn(c).run()
	}
}
