package ffmpegutil

import "log"

type HlsSegmentsParams struct {
	InputFile       string
	StartIndex      int
	TotalSegments   int
	SegmentDuration int
}

type Worker struct {
	CurrentEpisodeID int
	messageChan      chan HlsSegmentsParams
	WorkMap          map[int]bool
}

func NewWorker() *Worker {
	workMap := make(map[int]bool)
	return &Worker{
		messageChan:      make(chan HlsSegmentsParams),
		WorkMap:          workMap,
		CurrentEpisodeID: 0,
	}
}

func (w *Worker) Process(message HlsSegmentsParams) {
	log.Println("message1: ", message)
	if _, ok := w.WorkMap[message.StartIndex]; !ok {
		w.WorkMap[message.StartIndex] = true
		w.messageChan <- message
	}
}

// func (w *Worker) processMessage(message HlsSegmentsParams) {
// 	ffmpegStartIndex := message.StartIndex
// 	log.Println("Processing ffmpegStartIndex1", message.StartIndex)
// 	_, ok := w.WorkMap[message.StartIndex]
// 	if ok {
// 		for i := message.StartIndex; i < message.StartIndex+message.TotalSegments; i++ {
// 			_, ok := w.WorkMap[i]
// 			log.Println("Processing ffmpegStartIndex2", message.StartIndex)
// 			if !ok {
// 				ffmpegStartIndex = i
// 				break
// 			}
// 		}
// 	}

// 	log.Println("Processing ffmpegStartIndex3", message.StartIndex)

// 	GenerateHlsSegments(message.InputFile, ffmpegStartIndex, message.TotalSegments, message.SegmentDuration)

// 	for i := ffmpegStartIndex; i < ffmpegStartIndex+message.TotalSegments-1; i++ {
// 		w.WorkMap[i] = true
// 	}
// 	println(w.WorkMap)
// }
