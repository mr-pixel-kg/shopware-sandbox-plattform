package docker

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullTrackerReportsProgressAndCompletion(t *testing.T) {
	t.Parallel()

	tracker := NewPullTracker()
	imageID := "image-1"
	tracker.Start(imageID)

	stream := strings.NewReader(`
{"status":"Downloading","id":"layer-1","progressDetail":{"current":25,"total":100}}
{"status":"Extracting","id":"layer-1","progressDetail":{"current":50,"total":100}}
{"status":"Pull complete","id":"layer-1","progressDetail":{"current":0,"total":0}}
`)

	require.NoError(t, tracker.ConsumePullStream(imageID, stream))

	progress := tracker.Progress(imageID)
	assert.Equal(t, "pulling", progress.Status)
	assert.GreaterOrEqual(t, progress.Percent, 75)

	ch, cancel := tracker.Watch(imageID)
	defer cancel()
	tracker.Finish(imageID, nil)

	final, ok := <-ch
	require.True(t, ok)
	assert.Equal(t, "complete", final.Status)
	assert.Equal(t, 100, final.Percent)
}

func TestPullTrackerFinishWithErrorMarksFailure(t *testing.T) {
	t.Parallel()

	tracker := NewPullTracker()
	imageID := "image-2"
	tracker.Start(imageID)

	ch, cancel := tracker.Watch(imageID)
	defer cancel()
	tracker.Finish(imageID, assert.AnError)

	progress, ok := <-ch
	require.True(t, ok)
	assert.Equal(t, "failed", progress.Status)
	assert.Equal(t, assert.AnError.Error(), progress.Error)
	assert.Equal(t, 0, progress.Percent)
}
