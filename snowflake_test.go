package snowflake

import "testing"

func BenchmarkWorker_NextId(b *testing.B) {
	worker, _ := NewSnowflake(1)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			worker.NextId()
		}
	})
}

func TestWorker_NextId(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"test1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := NewSnowflake(1)
			for i := 0; i < 17000; i++ {
				id := w.NextId()
				t.Logf("NextId() got = %d", id)
			}
		})
	}
}
