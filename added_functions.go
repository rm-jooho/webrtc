package webrtc

import (
	"errors"

	"github.com/rm-jooho/webrtc/v3/pkg/rtcerr"
)

// jhms
func (pc *PeerConnection) HoldOnNegotiation() {
	if pc.negoHold == 0 {
		pc.negoHold = 1
	}
}

func (pc *PeerConnection) ResumeOnNegotiation() error {
	previousVal := pc.negoHold
	pc.negoHold = 0
	if pc.isClosed.get() {
		return &rtcerr.InvalidStateError{Err: ErrConnectionClosed}
	}
	if previousVal >= 1 {
		if previousVal > 1 {
			pc.onNegotiationNeeded()
			return nil
		}
	} else {
		// never called HoldOnNegotiation
		return errors.New("nothing to nego")
	}
	return nil
}

// jhms

func (m *MediaEngine) GetHeaderExtensionID(extension RTPHeaderExtensionCapability) (val int, audioNegotiated, videoNegotiated bool) {
	return m.getHeaderExtensionID(extension)
}

// AddTrack adds a Track to the PeerConnection
func (pc *PeerConnection) AddTrackSendonly(track TrackLocal, ssrc uint32) (*RTPSender, error) {
	if pc.isClosed.get() {
		return nil, &rtcerr.InvalidStateError{Err: ErrConnectionClosed}
	}

	var transceiver *RTPTransceiver
	/*
		for i, t := range pc.GetTransceivers() {
			fmt.Printf("idx:%d, stopped:%v kind:%s sender:%p direction:%s\n", i, t.stopped, t.kind.String(), t.Sender(), t.Direction().String())
		}
	*/
	for _, t := range pc.GetTransceivers() {
		if !t.stopped && t.kind == track.Kind() && (t.Sender() == nil && t.Direction() == RTPTransceiverDirectionInactive) {
			transceiver = t
			break
		}
	}
	if transceiver != nil {
		sender, err := pc.api.NewRTPSender(track, pc.dtlsTransport, SSRC(ssrc))
		if err != nil {
			return nil, err
		}
		transceiver.setSender(sender)
		// we still need to call setSendingTrack to ensure direction has changed
		if err := transceiver.setSendingTrack(track); err != nil {
			// inactive 상태였다면 setSendingTrack시 sendonly로 변함
			//transceiver.setDirection(RTPTransceiverDirectionSendonly)
			return nil, err
		}
		pc.onNegotiationNeeded()

		return sender, nil
	}

	var err error
	if ssrc == 0 {
		init := RtpTransceiverInit{
			Direction: RTPTransceiverDirectionSendonly,
		}
		transceiver, err = pc.AddTransceiverFromTrack(track, init)
	} else {
		init := RtpTransceiverInit{
			Direction: RTPTransceiverDirectionSendonly,
			SendEncodings: []RTPEncodingParameters{
				RTPEncodingParameters{
					RTPCodingParameters{
						SSRC: SSRC(ssrc),
					},
				},
			},
		}
		transceiver, err = pc.AddTransceiverFromTrack(track, init)
	}
	if err != nil {
		return nil, err
	}

	return transceiver.Sender(), nil
}
