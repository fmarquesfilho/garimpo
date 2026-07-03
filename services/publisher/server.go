package main

import (
	"context"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	publisherpb "github.com/fmarquesfilho/garimpo/gen/go/publisher/v1"
	"github.com/fmarquesfilho/garimpo/internal/publish"
)

// PublisherServer implementa publisher.v1.PublisherService.
type PublisherServer struct {
	publisherpb.UnimplementedPublisherServiceServer
	dispatcher *publish.Dispatcher
	destinos   *publish.MemDestinoStore
}

func NewPublisherServer() *PublisherServer {
	destinos := publish.NovoMemDestinoStore()

	var senders []publish.Sender
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		senders = append(senders, publish.NovoTelegramSender(token))
	}
	if ws := publish.NovoWhatsAppSenderFromEnv(); ws != nil {
		senders = append(senders, ws)
	}

	tipoPadrao := "telegram"
	configPadrao := os.Getenv("TELEGRAM_CHAT_ID")
	if os.Getenv("TELEGRAM_BOT_TOKEN") == "" {
		tipoPadrao = "whatsapp"
		configPadrao = ""
	}

	dispatcher := publish.NovoDispatcher(
		publish.DispatcherConfig{
			Destinos:     destinos,
			TipoPadrao:   tipoPadrao,
			ConfigPadrao: configPadrao,
		},
		senders...,
	)

	return &PublisherServer{
		dispatcher: dispatcher,
		destinos:   destinos,
	}
}

func (s *PublisherServer) Publish(ctx context.Context, req *publisherpb.PublishRequest) (*publisherpb.PublishResponse, error) {
	if req.GetContent() == nil {
		return nil, status.Error(codes.InvalidArgument, "content é obrigatório") //nolint:wrapcheck // gRPC status
	}

	c := req.GetContent()

	// Description contém a legenda customizada (HTML) quando o frontend edita.
	// Se for texto simples (sem tags HTML), não é legenda — é apenas a categoria.
	legendaHTML := ""
	if desc := c.GetDescription(); desc != "" && strings.Contains(desc, "<") {
		legendaHTML = desc
	}

	oferta := publish.Oferta{
		Nome:        c.GetTitle(),
		Preco:       c.GetPrice(),
		Link:        c.GetProductUrl(),
		Imagem:      c.GetImageUrl(),
		LegendaHTML: legendaHTML,
		DestinoID:   req.GetGroupId(),
		Estrategia:  "grpc",
	}

	res, err := s.dispatcher.Publicar(ctx, oferta)
	if err != nil {
		return &publisherpb.PublishResponse{
			Success:     false,
			PublishedAt: time.Now().UTC().Format(time.RFC3339),
		}, status.Errorf(codes.Internal, "falha ao publicar: %v", err)
	}

	return &publisherpb.PublishResponse{
		Success:     res.Enviado,
		MessageId:   res.SubID,
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *PublisherServer) ListGroups(ctx context.Context, req *publisherpb.ListGroupsRequest) (*publisherpb.ListGroupsResponse, error) {
	destinos, err := s.destinos.Listar(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao listar grupos: %v", err)
	}

	var groups []*publisherpb.Group
	for _, d := range destinos {
		if req.GetChannel() != "" && d.Tipo != req.GetChannel() {
			continue
		}
		groups = append(groups, &publisherpb.Group{
			Id:      d.ID,
			Name:    d.Nome,
			Channel: d.Tipo,
			Active:  d.Ativo,
		})
	}

	return &publisherpb.ListGroupsResponse{Groups: groups}, nil
}
