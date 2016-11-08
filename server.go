package socks5

import (
	"errors"
	"io"
	"log"
)

var (
	ERROR_VERSION           = errors.New("Invalid Version")
	ERROR_USER_PASS_VERSION = errors.New("Invalid Version of Username Password Auth")
	ERROR_BAD_REQUEST       = errors.New("Bad Request")
)

func NewNegotiationRequestFrom(r io.Reader) (*NegotiationRequest, error) {
	// memory strict
	bb := make([]byte, 2)
	if _, err := io.ReadFull(r, bb); err != nil {
		return nil, err
	}
	if bb[0] != VER {
		return nil, ERROR_VERSION
	}
	if bb[1] == 0 {
		return nil, ERROR_BAD_REQUEST
	}
	ms := make([]byte, int(bb[1]))
	if _, err := io.ReadFull(r, ms); err != nil {
		return nil, err
	}
	if Debug {
		log.Printf("Got NegotiationRequest: %#v %#v %#v\n", bb[0], bb[1], ms)
	}
	return &NegotiationRequest{
		Ver:      bb[0],
		NMethods: bb[1],
		Methods:  ms,
	}, nil
}

func NewNegotiationReply(method byte) *NegotiationReply {
	return &NegotiationReply{
		Ver:    VER,
		Method: method,
	}
}
func (r *NegotiationReply) WriteTo(w io.Writer) error {
	if _, err := w.Write([]byte{r.Ver, r.Method}); err != nil {
		return err
	}
	if Debug {
		log.Printf("Sent NegotiationReply: %#v %#v\n", r.Ver, r.Method)
	}
	return nil
}

func NewUserPassNegotiationRequestFrom(r io.Reader) (*UserPassNegotiationRequest, error) {
	bb := make([]byte, 2)
	if _, err := io.ReadFull(r, bb); err != nil {
		return nil, err
	}
	if bb[0] != USER_PASS_VER {
		return nil, ERROR_USER_PASS_VERSION
	}
	if bb[1] == 0 {
		return nil, ERROR_BAD_REQUEST
	}
	ub := make([]byte, int(bb[1])+1)
	if _, err := io.ReadFull(r, ub); err != nil {
		return nil, err
	}
	if ub[int(bb[1])] == 0 {
		return nil, ERROR_BAD_REQUEST
	}
	p := make([]byte, int(ub[int(bb[1])]))
	if _, err := io.ReadFull(r, p); err != nil {
		return nil, err
	}
	if Debug {
		log.Printf("Got UserPassNegotiationRequest: %#v %#v %#v %#v %#v\n", bb[0], bb[1], ub[:int(bb[1])], ub[int(bb[1])], p)
	}
	return &UserPassNegotiationRequest{
		Ver:    bb[0],
		Ulen:   bb[1],
		Uname:  ub[:int(bb[1])],
		Plen:   ub[int(bb[1])],
		Passwd: p,
	}, nil
}

func NewUserPassNegotiationReply(status byte) *UserPassNegotiationReply {
	return &UserPassNegotiationReply{
		Ver:    USER_PASS_VER,
		Status: status,
	}
}

func (r *UserPassNegotiationReply) WriteTo(w io.Writer) error {
	if _, err := w.Write([]byte{r.Ver, r.Status}); err != nil {
		return err
	}
	if Debug {
		log.Printf("Sent UserPassNegotiationReply: %#v %#v \n", r.Ver, r.Status)
	}
	return nil
}

func NewRequestFrom(r io.Reader) (*Request, error) {
	bb := make([]byte, 4)
	if _, err := io.ReadFull(r, bb); err != nil {
		return nil, err
	}
	if bb[0] != VER {
		return nil, ERROR_VERSION
	}
	var addr []byte
	if bb[3] == ATYP_IPV4 {
		addr = make([]byte, 4)
		if _, err := io.ReadFull(r, addr); err != nil {
			return nil, err
		}
	} else if bb[3] == ATYP_IPV6 {
		addr = make([]byte, 16)
		if _, err := io.ReadFull(r, addr); err != nil {
			return nil, err
		}
	} else if bb[3] == ATYP_DOMAIN {
		dal := make([]byte, 1)
		if _, err := io.ReadFull(r, dal); err != nil {
			return nil, err
		}
		if dal[0] == 0 {
			return nil, ERROR_BAD_REQUEST
		}
		addr = make([]byte, int(dal[0]))
		if _, err := io.ReadFull(r, addr); err != nil {
			return nil, err
		}
		addr = append(dal, addr...)
	} else {
		return nil, ERROR_BAD_REQUEST
	}
	port := make([]byte, 2)
	if _, err := io.ReadFull(r, port); err != nil {
		return nil, err
	}
	if Debug {
		log.Printf("Got Request: %#v %#v %#v %#v %#v %#v\n", bb[0], bb[1], bb[2], bb[3], addr, port)
	}
	return &Request{
		Ver:     bb[0],
		Cmd:     bb[1],
		Rsv:     bb[2],
		Atyp:    bb[3],
		DstAddr: addr,
		DstPort: port,
	}, nil
}

func NewReply(rep byte, atyp byte, bndaddr []byte, bndport []byte) *Reply {
	if atyp == ATYP_DOMAIN {
		bndaddr = append([]byte{byte(len(bndaddr))}, bndaddr...)
	}
	return &Reply{
		Ver:     VER,
		Rep:     rep,
		Rsv:     0x00,
		Atyp:    atyp,
		BndAddr: bndaddr,
		BndPort: bndport,
	}
}

func (r *Reply) WriteTo(w io.Writer) error {
	if _, err := w.Write([]byte{r.Ver, r.Rep, r.Rsv, r.Atyp}); err != nil {
		return err
	}
	if _, err := w.Write(r.BndAddr); err != nil {
		return err
	}
	if _, err := w.Write(r.BndPort); err != nil {
		return err
	}
	if Debug {
		log.Printf("Sent Reply: %#v %#v %#v %#v %#v %#v\n", r.Ver, r.Rep, r.Rsv, r.Atyp, r.BndAddr, r.BndPort)
	}
	return nil
}
