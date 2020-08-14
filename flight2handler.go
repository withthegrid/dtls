package dtls

import (
	"bytes"
	"context"
)

func flight2Parse(ctx context.Context, c flightConn, state *State, cache *handshakeCache, cfg *handshakeConfig) (flightVal, *alert, error) {
	seq, msgs, ok := cache.fullPullMap(state.handshakeRecvSequence,
		handshakeCachePullRule{handshakeTypeClientHello, cfg.initialEpoch, true, false},
	)
	if !ok {
		// Client may retransmit the first ClientHello when HelloVerifyRequest is dropped.
		// Parse as flight 0 in this case.
		cfg.log.Tracef("[flight2Parse] -> flight0Parse")
		return flight0Parse(ctx, c, state, cache, cfg)
	}
	state.handshakeRecvSequence = seq

	var clientHello *handshakeMessageClientHello

	// Validate type
	if clientHello, ok = msgs[handshakeTypeClientHello].(*handshakeMessageClientHello); !ok {
		cfg.log.Tracef("[flight2Parse] clientHello !ok alertInternalError")
		return 0, &alert{alertLevelFatal, alertInternalError}, nil
	}

	if !clientHello.version.Equal(protocolVersion1_2) {
		cfg.log.Tracef("[flight2Parse] clientHello alertProtocolVersion")
		return 0, &alert{alertLevelFatal, alertProtocolVersion}, errUnsupportedProtocolVersion
	}

	if len(clientHello.cookie) == 0 {
		cfg.log.Tracef("[flight2Parse] clientHello.cookie empty")
		return 0, nil, nil
	}
	if !bytes.Equal(state.cookie, clientHello.cookie) {
		cfg.log.Tracef("[flight2Parse] clientHello.cookie mismatch")
		return 0, &alert{alertLevelFatal, alertAccessDenied}, errCookieMismatch
	}

	cfg.log.Tracef("[flight2Parse] -> flight4")
	return flight4, nil, nil
}

func flight2Generate(c flightConn, state *State, cache *handshakeCache, cfg *handshakeConfig) ([]*packet, *alert, error) {
	state.handshakeSendSequence = 0
	return []*packet{
		{
			record: &recordLayer{
				recordLayerHeader: recordLayerHeader{
					protocolVersion: protocolVersion1_2,
				},
				content: &handshake{
					handshakeMessage: &handshakeMessageHelloVerifyRequest{
						version: protocolVersion1_2,
						cookie:  state.cookie,
					},
				},
			},
		},
	}, nil, nil
}
