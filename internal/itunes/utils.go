package itunes

import (
	"errors"
	//"log"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func GetCurrentTrack(dispatcher *ole.IDispatch) (*IiTrack, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is not ready")
	}
	//log.Printf("called GetCurrentTrack")
	var err error = nil
	trackProp, err := oleutil.GetProperty(dispatcher, "CurrentTrack")
	track, err := getCOMObjectFromVariant[IiTrack](trackProp, IID_IiTrack)
	return track, err
}

func GetCurrentTunes(dispatcher *ole.IDispatch) (*IiTunes, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is not ready")
	}
	
	var err error = nil
	//log.Printf("called GetCurrentTunes")
	
	soundVolumeVar, err := oleutil.GetProperty(dispatcher, "SoundVolume")
	soundVolume := int(soundVolumeVar.Val)
	
	playerPositionVar, err := oleutil.GetProperty(dispatcher, "PlayerPosition")
	playerPosition := int(playerPositionVar.Val)
	
	playerPositionMSVar, err := oleutil.GetProperty(dispatcher, "PlayerPositionMS")
	playerPositionMS := int(playerPositionMSVar.Val)
	
	playerStateVar, err := oleutil.GetProperty(dispatcher, "PlayerState")
	playerState := int(playerStateVar.Val)
	
	tunes := &IiTunes{
		SoundVolume: int64(soundVolume),
		PlayerPosition: int32(playerPosition),
		PlayerPositionMS: int64(playerPositionMS),
		PlayerState: int32(playerState),
	}

	//log.Printf("tunes object: %v", tunes)

	return tunes, err
}

// func GetPlayerButtonsState(dispatcher *ole.IDispatch) {
// 	buttonsState := oleutil.CallMethod()
// }