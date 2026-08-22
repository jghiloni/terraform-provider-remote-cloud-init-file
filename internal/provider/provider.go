package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type WriteFileProvider struct {
	version string
}

// Configure implements [provider.Provider].
func (w *WriteFileProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

// DataSources implements [provider.Provider].
func (w *WriteFileProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

// Metadata implements [provider.Provider].
func (w *WriteFileProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "remote-cloud-init-file"
	resp.Version = w.version
}

// Resources implements [provider.Provider].
func (w *WriteFileProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWriteFileResource,
	}
}

// Schema implements [provider.Provider].
func (w *WriteFileProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "A provider for creating cloud-init write_file entries",
		MarkdownDescription: "This provider has no configurations",
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WriteFileProvider{
			version,
		}
	}
}

var _ provider.Provider = (*WriteFileProvider)(nil)
