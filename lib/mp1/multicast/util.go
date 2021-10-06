package multicast

import "fmt"

func MaxUint64(x uint64, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

func MaxOfArrayUint64(arr []uint64) (uint64, error) {
	if len(arr) < 1 {
		return 0, fmt.Errorf("arr is an empty sequence")
	}

	max := arr[0]
	for _, value := range arr {
		if max < value {
			max = value
		}
	}
	return max, nil
}

func MaxOfArrayProposalItem(arr []*ProposalItem) (*ProposalItem, error) {
	if len(arr) == 0 {
		return nil, fmt.Errorf("arr is an empty sequence")
	}

	max := arr[0]
	for _, value := range arr {
		if max.ProposalSeqNum < value.ProposalSeqNum {
			max = value
			continue
		}
		if max.ProposalSeqNum > value.ProposalSeqNum {
			continue
		}
		if max.ProposalSeqNum == value.ProposalSeqNum {
			if max.ProcessID < value.ProcessID {
				max = value
			}
		}
	}
	return max, nil
}
