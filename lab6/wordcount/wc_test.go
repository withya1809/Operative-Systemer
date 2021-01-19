package wordcount

import (
	"testing"
)

func TestWordCount(t *testing.T) {
	b := loadMoby()
	cnt := wordCount(b)
	expectedCnt := 175131
	if cnt != expectedCnt {
		t.Errorf("expected %d words, found %d words\n", expectedCnt, cnt)
	}
}

func TestParallelWordCount(t *testing.T) {
	b := loadMoby()
	cnt := parallelWordCount(b)
	expectedCnt := 175131
	if cnt != expectedCnt {
		t.Errorf("expected %d words, found %d words\n", expectedCnt, cnt)
	}
}

func TestParallelWordCountManyShards(t *testing.T) {
	b := loadMoby()
	wc := wordCount(b)
	for shards := 1; shards < 10; shards++ {
		pwc := doParallelWordCount(b, shards)
		if pwc != wc {
			t.Errorf("doParallelWordCount(..., %d)=%d, expected %d", shards, pwc, wc)
		}
	}
}

func BenchmarkWordCountSequential(b *testing.B) {
	mobyDick := loadMoby()
	for i := 0; i < b.N; i++ {
		_ = wordCount(mobyDick)
	}
}

func BenchmarkWordCountParallel(b *testing.B) {
	mobyDick := loadMoby()
	for i := 0; i < b.N; i++ {
		_ = parallelWordCount(mobyDick)
	}
}
