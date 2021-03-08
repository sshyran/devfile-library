package v2

import (
	"fmt"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// AddVolumeMounts adds the volume mounts to the specified container component
func (d *DevfileV2) AddVolumeMounts(componentName string, volumeMounts []v1.VolumeMount) error {
	var pathErrorContainers []string
	found := false
	for _, component := range d.Components {
		if component.Container != nil && component.Name == componentName {
			found = true
			for _, devfileVolumeMount := range component.Container.VolumeMounts {
				for _, volumeMount := range volumeMounts {
					if devfileVolumeMount.Path == volumeMount.Path {
						pathErrorContainers = append(pathErrorContainers, fmt.Sprintf("unable to mount volume %s, as another volume %s is mounted to the same path %s in the container %s", volumeMount.Name, devfileVolumeMount.Name, volumeMount.Path, component.Name))
					}
				}
			}
			if len(pathErrorContainers) == 0 {
				component.Container.VolumeMounts = append(component.Container.VolumeMounts, volumeMounts...)
			}
		}
	}

	if !found {
		return &common.FieldNotFoundError{
			Field: "container component",
			Name:  componentName,
		}
	}

	if len(pathErrorContainers) > 0 {
		return fmt.Errorf("errors while adding volume mounts:\n%s", strings.Join(pathErrorContainers, "\n"))
	}

	return nil
}

// DeleteVolumeMount deletes the volume mount from container components
func (d *DevfileV2) DeleteVolumeMount(name string) error {
	found := false
	for i := range d.Components {
		if d.Components[i].Container != nil && d.Components[i].Name != name {
			// Volume Mounts can have multiple instances of a volume mounted at different paths
			// As arrays are rearraged/shifted for deletion, we lose one element every time there is a match
			// Looping backward is efficient, otherwise we would have to manually decrement counter
			// if we looped forward
			for j := len(d.Components[i].Container.VolumeMounts) - 1; j >= 0; j-- {
				if d.Components[i].Container.VolumeMounts[j].Name == name {
					found = true
					d.Components[i].Container.VolumeMounts = append(d.Components[i].Container.VolumeMounts[:j], d.Components[i].Container.VolumeMounts[j+1:]...)
				}
			}
		}
	}

	if !found {
		return &common.FieldNotFoundError{
			Field: "volume mount",
			Name:  name,
		}
	}

	return nil
}

// GetVolumeMountPath gets the mount path of the specified volume mount from the specified container component
func (d *DevfileV2) GetVolumeMountPath(mountName, componentName string) (string, error) {
	componentFound := false

	for _, component := range d.Components {
		if component.Container != nil && component.Name == componentName {
			componentFound = true
			for _, volumeMount := range component.Container.VolumeMounts {
				if volumeMount.Name == mountName {
					return volumeMount.Path, nil
				}
			}
		}
	}

	if !componentFound {
		return "", &common.FieldNotFoundError{
			Field: "container component",
			Name:  componentName,
		}
	}

	return "", fmt.Errorf("volume %s not mounted to component %s", mountName, componentName)
}
