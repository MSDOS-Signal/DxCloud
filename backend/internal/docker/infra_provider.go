package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
)

// ---------- ImageProvider ----------

func (d *DockerProvider) ListImages(ctx context.Context) ([]ImageInfo, error) {
	list, err := d.cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	out := make([]ImageInfo, 0, len(list))
	for _, s := range list {
		created := time.Unix(s.Created, 0)
		info := ImageInfo{
			ID:          s.ID,
			RepoTags:    s.RepoTags,
			RepoDigests: s.RepoDigests,
			Size:        s.Size,
			CreatedAt:   created,
		}
		out = append(out, info)
	}
	return out, nil
}

func (d *DockerProvider) RemoveImage(ctx context.Context, ref string, force bool) error {
	_, err := d.cli.ImageRemove(ctx, ref, image.RemoveOptions{Force: force, PruneChildren: true})
	return err
}

func (d *DockerProvider) TagImage(ctx context.Context, src, dst string) error {
	return d.cli.ImageTag(ctx, src, dst)
}

func (d *DockerProvider) InspectImage(ctx context.Context, ref string) (*ImageInfo, error) {
	insp, _, err := d.cli.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}
	info := &ImageInfo{
		ID:          insp.ID,
		RepoTags:    insp.RepoTags,
		RepoDigests: insp.RepoDigests,
		Size:        insp.Size,
	}
	if t, err := time.Parse(time.RFC3339Nano, insp.Created); err == nil {
		info.CreatedAt = t
	}
	return info, nil
}

// ---------- NetworkProvider ----------

func (d *DockerProvider) CreateNetwork(ctx context.Context, spec NetworkSpec) (NetworkInfo, error) {
	if spec.Driver == "" {
		spec.Driver = "bridge"
	}
	create := types.NetworkCreate{
		Driver:   spec.Driver,
		Internal: spec.Internal,
		Labels:   spec.Labels,
	}
	if spec.Subnet != "" || spec.Gateway != "" || spec.IPRange != "" {
		create.IPAM = &network.IPAM{
			Config: []network.IPAMConfig{{
				Subnet:  spec.Subnet,
				Gateway: spec.Gateway,
				IPRange: spec.IPRange,
			}},
		}
	}
	resp, err := d.cli.NetworkCreate(ctx, spec.Name, create)
	if err != nil {
		return NetworkInfo{}, err
	}
	info, err := d.InspectNetwork(ctx, resp.ID)
	if err != nil {
		return NetworkInfo{}, err
	}
	info.ID = resp.ID
	return *info, nil
}

func (d *DockerProvider) RemoveNetwork(ctx context.Context, id string) error {
	return d.cli.NetworkRemove(ctx, id)
}

func (d *DockerProvider) InspectNetwork(ctx context.Context, id string) (*NetworkInfo, error) {
	res, err := d.cli.NetworkInspect(ctx, id, types.NetworkInspectOptions{})
	if err != nil {
		return nil, err
	}
	return networkInfoFromResource(&res), nil
}

func (d *DockerProvider) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	list, err := d.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]NetworkInfo, 0, len(list))
	for i := range list {
		out = append(out, *networkInfoFromResource(&list[i]))
	}
	return out, nil
}

func networkInfoFromResource(res *types.NetworkResource) *NetworkInfo {
	info := &NetworkInfo{
		ID:         res.ID,
		Name:       res.Name,
		Driver:     res.Driver,
		Containers: map[string]string{},
	}
	if res.IPAM.Config != nil && len(res.IPAM.Config) > 0 {
		info.Subnet = res.IPAM.Config[0].Subnet
		info.Gateway = res.IPAM.Config[0].Gateway
		info.IPRange = res.IPAM.Config[0].IPRange
	}
	for _, ep := range res.Containers {
		info.Containers[ep.Name] = ep.IPv4Address
	}
	return info
}

func (d *DockerProvider) ConnectNetwork(ctx context.Context, netID, containerID, fixedIP string) error {
	ep := &network.EndpointSettings{}
	if fixedIP != "" {
		ep.IPAMConfig = &network.EndpointIPAMConfig{IPv4Address: fixedIP}
	}
	return d.cli.NetworkConnect(ctx, netID, containerID, ep)
}

func (d *DockerProvider) DisconnectNetwork(ctx context.Context, netID, containerID string) error {
	return d.cli.NetworkDisconnect(ctx, netID, containerID, true)
}

// ---------- StorageProvider ----------

func (d *DockerProvider) CreateVolume(ctx context.Context, name string, labels map[string]string) (VolumeInfo, error) {
	v, err := d.cli.VolumeCreate(ctx, volume.CreateOptions{Name: name, Labels: labels})
	if err != nil {
		return VolumeInfo{}, err
	}
	return VolumeInfo{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, Labels: v.Labels}, nil
}

func (d *DockerProvider) RemoveVolume(ctx context.Context, name string) error {
	return d.cli.VolumeRemove(ctx, name, true)
}

func (d *DockerProvider) InspectVolume(ctx context.Context, name string) (*VolumeInfo, error) {
	v, err := d.cli.VolumeInspect(ctx, name)
	if err != nil {
		return nil, err
	}
	return &VolumeInfo{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, Labels: v.Labels}, nil
}

func (d *DockerProvider) ListVolumes(ctx context.Context) ([]VolumeInfo, error) {
	resp, err := d.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]VolumeInfo, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		out = append(out, VolumeInfo{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, Labels: v.Labels})
	}
	return out, nil
}

// 断言 DockerProvider 实现聚合 Provider 接口。
var _ Provider = (*DockerProvider)(nil)
