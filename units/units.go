package units

type Server struct {
	Name string
}

type Client struct {
	Name   string
	Server *Server
}
