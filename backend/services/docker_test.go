package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupDockerService(t *testing.T) *DockerService {
	dockerService, err := NewDockerService()
	assert.NoError(t, err, "Creating docker service should not return an error")
	return dockerService
}

func TestListImages(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("ListImages should return images", func(t *testing.T) {
		images, err := dockerService.ListImages(context.Background())

		assert.NoError(t, err, "ListImages should not return an error")
		assert.NotEmpty(t, images, "ListImages should return at least one image")

		for _, img := range images {
			assert.NotEmpty(t, img.ID, "Image ID should not be empty")
			assert.NotEmpty(t, img.Name, "Image Name should not be empty")
			assert.NotEmpty(t, img.Tag, "Image Tag should not be empty")
			assert.False(t, img.Created.IsZero(), "Created timestamp should be valid")
			assert.Greater(t, img.Size, int64(0), "Image size should be greater than 0")
		}
	})
}

func TestGetImage(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("GetImage should return nginx:latest", func(t *testing.T) {
		img, err := dockerService.GetImage(context.Background(), "nginx:latest")

		assert.NoError(t, err, "GetImage should not return an error")
		assert.NotEmpty(t, img, "Image should not be empty")
		assert.Equal(t, "nginx", img.Name, "Image Name should be nginx")
		assert.Equal(t, "latest", img.Tag, "Image Tag should be latest")
		assert.False(t, img.Created.IsZero(), "Created timestamp should be valid")
		assert.Greater(t, img.Size, int64(0), "Image size should be greater than 0")
	})

	t.Run("GetImage should throw error when image does not exist", func(t *testing.T) {
		img, err := dockerService.GetImage(context.Background(), "nonexistant:latest")

		assert.Error(t, err, "GetImage should return an error")
		assert.Empty(t, img, "Image should be empty")
	})
}

func TestPullImage(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("PullImage should pull alpine:latest", func(t *testing.T) {
		img, err := dockerService.PullImage(context.Background(), "alpine:latest")

		assert.NoError(t, err, "PullImage should not return an error")
		assert.NotEmpty(t, img.ID, "Image ID should not be empty")
		assert.Equal(t, "alpine", img.Name, "Image Name should be alpine")
		assert.Equal(t, "latest", img.Tag, "Image Tag should be latest")
		assert.False(t, img.Created.IsZero(), "Created timestamp should be valid")
		assert.Greater(t, img.Size, int64(0), "Image size should be greater than 0")
	})

	t.Run("PullImage should throw no error when pulling image twice", func(t *testing.T) {
		img, err := dockerService.PullImage(context.Background(), "alpine:latest")

		assert.NoError(t, err, "PullImage should not return an error")
		assert.NotEmpty(t, img.ID, "Image ID should not be empty")
		assert.Equal(t, "alpine", img.Name, "Image Name should be alpine")
		assert.Equal(t, "latest", img.Tag, "Image Tag should be latest")
		assert.False(t, img.Created.IsZero(), "Created timestamp should be valid")
		assert.Greater(t, img.Size, int64(0), "Image size should be greater than 0")
	})
}

func TestRemoveImage(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("RemoveImage should remove alpine:latest", func(t *testing.T) {
		err := dockerService.RemoveImage(context.Background(), "alpine:latest")
		assert.NoError(t, err, "Removing docker image should not return an error")
	})

	t.Run("RemoveImage should return error when deleting non-existing image", func(t *testing.T) {
		err := dockerService.RemoveImage(context.Background(), "alpine:latest")
		assert.Error(t, err, "Removing docker image should return an error")
	})
}

func TestListContainers(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("ListContainers should return containers", func(t *testing.T) {
		containers, err := dockerService.ListContainers(context.Background())

		assert.NoError(t, err, "ListContainers should not return an error")
		assert.NotEmpty(t, containers, "ListContainers should return at least one container")

		for _, cont := range containers {
			assert.NotEmpty(t, cont.ID, "Container ID should not be empty")
			assert.NotEmpty(t, cont.Name, "Container Name should not be empty")
			assert.NotEmpty(t, cont.Image, "Container Image should not be empty")
			assert.False(t, cont.Created.IsZero(), "Created timestamp should be valid")
			assert.NotEmpty(t, cont.Status, "Container Status should not be empty")
			assert.NotNil(t, cont.Labels, "Container Labels should not be nil")
		}
	})
}

func TestGetContainer(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("GetContainer should return a container", func(t *testing.T) {
		cont, err := dockerService.GetContainer(context.Background(), "shopware_6.6")

		assert.NoError(t, err, "GetContainer should not return an error")
		assert.NotEmpty(t, cont, "Container should not be empty")

		assert.NotEmpty(t, cont.ID, "Container ID should not be empty")
		assert.NotEmpty(t, cont.Name, "Container Name should not be empty")
		assert.NotEmpty(t, cont.Image, "Container Image should not be empty")
		assert.False(t, cont.Created.IsZero(), "Created timestamp should be valid")
		assert.NotEmpty(t, cont.Status, "Container Status should not be empty")
		assert.NotNil(t, cont.Labels, "Container Labels should not be nil")
	})

	t.Run("GetContainer should throw error when container does not exist", func(t *testing.T) {
		cont, err := dockerService.GetContainer(context.Background(), "nonexistant")

		assert.Error(t, err, "GetContainer should return an error")
		assert.Empty(t, cont, "Container should be empty")
	})
}

func TestCreateContainer(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("CreateContainer should throw an error if image does not exists", func(t *testing.T) {
		cont, err := dockerService.CreateContainer(context.Background(), "nonexistant:latest", "test", nil)

		assert.Error(t, err, "CreateContainer should return an error")
		assert.Empty(t, cont, "Container should be empty")
	})

	t.Run("CreateContainer should create a container", func(t *testing.T) {
		// Pull image
		img, err := dockerService.PullImage(context.Background(), "alpine:latest")
		assert.NoError(t, err, "PullImage should not return an error")
		assert.NotEmpty(t, img.ID, "Image ID should not be empty")
		defer dockerService.RemoveImage(context.Background(), "alpine:latest")

		// Create container
		contId, err := dockerService.CreateContainer(context.Background(), "alpine:latest", "test", nil)

		assert.NoError(t, err, "CreateContainer should not return an error")
		assert.NotEmpty(t, contId, "ContainerId should not be empty")

		cont, err2 := dockerService.GetContainer(context.Background(), contId)
		assert.NoError(t, err2, "GetContainer should not return an error")

		assert.NotEmpty(t, cont.ID, "Container ID should not be empty")
		assert.NotEmpty(t, cont.Name, "Container Name should not be empty")
		assert.NotEmpty(t, cont.Image, "Container Image should not be empty")
		assert.False(t, cont.Created.IsZero(), "Created timestamp should be valid")
		assert.NotEmpty(t, cont.Status, "Container Status should not be empty")
		assert.NotNil(t, cont.Labels, "Container Labels should not be nil")
	})

}

func TestStartContainer(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("StartContainer should throw an error if container does not exists", func(t *testing.T) {
		err := dockerService.StartContainer(context.Background(), "nonExistingContainer")

		assert.Error(t, err, "StartContainer should return an error")
	})

	t.Run("StartContainer should start the container", func(t *testing.T) {
		err := dockerService.StartContainer(context.Background(), "test")

		assert.NoError(t, err, "StartContainer should not return an error")
	})
}

func TestStopContainer(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("StopContainer should throw an error if container does not exists", func(t *testing.T) {
		err := dockerService.StartContainer(context.Background(), "nonExistingContainer")

		assert.Error(t, err, "StopContainer should return an error")
	})

	t.Run("StopContainer should stop the container", func(t *testing.T) {
		err := dockerService.StartContainer(context.Background(), "test")

		assert.NoError(t, err, "StopContainer should not return an error")
	})
}

func TestRemoveContainer(t *testing.T) {
	dockerService := setupDockerService(t)

	t.Run("RemoveContainer should remove container", func(t *testing.T) {
		err := dockerService.RemoveContainer(context.Background(), "test")
		assert.NoError(t, err, "Removing docker container should not return an error")
	})

	t.Run("RemoveContainer should return error when deleting non-existing container", func(t *testing.T) {
		err := dockerService.RemoveContainer(context.Background(), "test")
		assert.Error(t, err, "Removing docker container should return an error")
	})
}
