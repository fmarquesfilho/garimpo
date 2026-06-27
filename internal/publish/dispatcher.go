package publish

import (
	"context"
	"fmt"
)

// Dispatcher é o publicador principal: roteia a oferta para o Sender correto
// baseado no destino. Implementa Publicador.
//
// Padrão: Registry + Strategy. Os Senders se registram por tipo; o Dispatcher
// resolve qual usar em runtime baseado no Destino selecionado.
type Dispatcher struct {
	senders  map[string]Sender // tipo → sender (ex.: "telegram" → TelegramSender)
	destinos DestinoStore      // onde os destinos estão persistidos
	padrao   string            // config padrão (chat_id/telefone) se DestinoID estiver vazio
	tipoPad  string            // tipo padrão (ex.: "telegram")
}

// DispatcherConfig configura o Dispatcher.
type DispatcherConfig struct {
	Destinos     DestinoStore
	TipoPadrao   string // provedor padrão quando DestinoID não é informado
	ConfigPadrao string // config do destino padrão (ex.: chat_id da env)
}

// NovoDispatcher cria o dispatcher com os senders registrados.
func NovoDispatcher(cfg DispatcherConfig, senders ...Sender) *Dispatcher {
	d := &Dispatcher{
		senders:  make(map[string]Sender, len(senders)),
		destinos: cfg.Destinos,
		padrao:   cfg.ConfigPadrao,
		tipoPad:  cfg.TipoPadrao,
	}
	for _, s := range senders {
		d.senders[s.Tipo()] = s
	}
	return d
}

func (d *Dispatcher) Nome() string { return "dispatcher" }

func (d *Dispatcher) Publicar(ctx context.Context, o Oferta) (Resultado, error) {
	// Se não especificou destino, usa o padrão (config da env)
	if o.DestinoID == "" {
		if d.padrao == "" {
			return Resultado{Canal: d.tipoPad, Enviado: false, Detalhe: "escolha um destino"},
				fmt.Errorf("nenhum destino selecionado e TELEGRAM_CHAT_ID não configurado")
		}
		sender, ok := d.senders[d.tipoPad]
		if !ok {
			return Resultado{Enviado: false, Detalhe: "provedor padrão não configurado"},
				fmt.Errorf("provedor %q não registrado", d.tipoPad)
		}
		return sender.Enviar(ctx, o, d.padrao)
	}

	// Busca o destino no store
	destino, err := d.destinos.Buscar(ctx, o.DestinoID)
	if err != nil {
		return Resultado{Enviado: false, Detalhe: err.Error()}, err
	}
	if !destino.Ativo {
		return Resultado{Enviado: false, Detalhe: "destino inativo"},
			fmt.Errorf("destino %q está inativo", o.DestinoID)
	}

	sender, ok := d.senders[destino.Tipo]
	if !ok {
		return Resultado{Enviado: false, Detalhe: "provedor não suportado"},
			fmt.Errorf("provedor %q não registrado", destino.Tipo)
	}

	return sender.Enviar(ctx, o, destino.Config)
}
