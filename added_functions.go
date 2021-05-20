package webrtc

import (
	"errors"

	"github.com/rm-jooho/webrtc/v3/pkg/rtcerr"
)

// 일시적으로 OnNegotiation 호출을 중지
// 복수의 로컬 미디어 처리를 한번의 SDP 생성으로 축약하기 위해 사용
// 충분히 테스트되지 않았음. 원하는 동작을 위해 부가 작업 필요함...개선 필요
func (pc *PeerConnection) HoldOnNegotiation() {
	if pc.negoHold == 0 {
		pc.negoHold = 1
	}
}

// HoldOnNegotiation에 의해 중지된 OnNegotiation 호출을 재개
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

// 사라진 GetHeaderExtensionID API
func (m *MediaEngine) GetHeaderExtensionID(extension RTPHeaderExtensionCapability) (val int, audioNegotiated, videoNegotiated bool) {
	return m.getHeaderExtensionID(extension)
}

// transceiver 재사용이 되고 sendonly만 적용되고 SSRC도 적용 가능한 transceiver 획득
func (pc *PeerConnection) AddSendonlyTrack(track TrackLocal, ssrc uint32) (*RTPSender, error) {
	if pc.isClosed.get() {
		return nil, &rtcerr.InvalidStateError{Err: ErrConnectionClosed}
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()
	for _, t := range pc.rtpTransceivers {
		if !t.stopped && t.kind == track.Kind() && t.Sender() == nil && t.Direction() == RTPTransceiverDirectionInactive {
			sender, err := pc.api.NewRTPSender(track, pc.dtlsTransport, SSRC(ssrc))
			if err == nil {
				err = t.SetSender(sender, track)
				if err != nil {
					_ = sender.Stop()
					t.setSender(nil)
				}
			}
			if err != nil {
				return nil, err
			}
			pc.onNegotiationNeeded()
			return sender, nil
		}
	}

	transceiver, err := pc.newTransceiverFromTrack(RTPTransceiverDirectionSendonly, track, SSRC(ssrc))
	if err != nil {
		return nil, err
	}
	pc.addRTPTransceiver(transceiver)
	return transceiver.Sender(), nil
}
