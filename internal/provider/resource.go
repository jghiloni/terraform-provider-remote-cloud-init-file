package provider

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sclevine/yj/convert"
	"go.yaml.in/yaml/v4"
)

type WriteFileResource struct{}

type resourceModel struct {
	// configurable parameters
	Encoding     types.String `tfsdk:"encoding"`
	Path         types.String `tfsdk:"path"`
	User         types.String `tfsdk:"user"`
	Group        types.String `tfsdk:"group"`
	Permissions  types.String `tfsdk:"permissions"`
	Defer        types.Bool   `tfsdk:"defer"`
	EncodeAs     types.String `tfsdk:"encode_as"`
	FileContents types.String `tfsdk:"file_contents"`

	// computed parameters
	ID              types.String `tfsdk:"id"`
	Content         types.String `tfsdk:"content"`
	ContentChecksum types.String `tfsdk:"content_sha256"`
}

func (w *WriteFileResource) upsert(ctx context.Context, plan tfsdk.Plan, state tfsdk.State, diagnostics diag.Diagnostics) {
	var data resourceModel
	diagnostics.Append(plan.Get(ctx, &data)...)

	model := WriteFile{
		Path:        data.Path.ValueString(),
		Owner:       fmt.Sprintf("%s:%s", data.User.ValueString(), data.Group.ValueString()),
		Permissions: data.Permissions.ValueString(),
		Defer:       true,
	}

	templateType := data.EncodeAs.ValueString()
	content := data.FileContents.ValueString()
	if templateType != "" {
		from := convert.HCL{}
		var to convert.Encoding
		switch templateType {
		case "json":
			to = convert.JSON{}
		case "yaml":
			to = convert.YAML{}
		}

		rep, err := from.Decode(strings.NewReader(content))
		if err != nil {
			diagnostics.AddAttributeError(path.Root("file_contents"), "could not decode file_contents as HCL", err.Error())
			return
		}

		w := new(strings.Builder)

		err = to.Encode(w, rep)
		if err != nil {
			diagnostics.AddAttributeError(path.Root("file_contents"), fmt.Sprintf("could not encode file_contents as %s", to), err.Error())
			return
		}

		content = w.String()
	}

	encoding := ""
	if !data.Encoding.IsNull() {
		encoding = data.Encoding.ValueString()
		model.Encoding = encoding
	}

	model.SetContents(content)

	modelYAML := new(bytes.Buffer)
	err := yaml.NewEncoder(modelYAML).Encode(model)
	if err != nil {
		diagnostics.AddError("could not write yaml", err.Error())
	}

	yamlBytes := modelYAML.Bytes()
	data.Content = types.StringValue(string(yamlBytes))

	idSum := sha1.New()
	data.ID = types.StringValue(hex.EncodeToString(idSum.Sum(yamlBytes)))

	check := sha256.New()
	data.ContentChecksum = types.StringValue(hex.EncodeToString(check.Sum(yamlBytes)))

	diagnostics.Append(state.Set(ctx, &data)...)
}

// Create implements [resource.Resource].
func (w *WriteFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	w.upsert(ctx, req.Plan, resp.State, resp.Diagnostics)
}

// Delete implements [resource.Resource].
func (w *WriteFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, new(resourceModel))...)
}

// Update implements [resource.Resource].
func (w *WriteFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	w.upsert(ctx, req.Plan, resp.State, resp.Diagnostics)
}

func (w *WriteFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName
}

func (w *WriteFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(req.State.Get(ctx, new(resourceModel))...)
}

func (w *WriteFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generate a yaml block suitable for a write_files entry in cloud-init",
		Attributes: map[string]schema.Attribute{
			"encoding": schema.StringAttribute{
				Optional:            true,
				Description:         "If set to b64, content will be base64 encoded. If set to gzip, it will be gzipped then base64 encoded",
				MarkdownDescription: "If set to b64, content will be base64 encoded. If set to gzip, it will be gzipped then base64 encoded",
				Validators:          []validator.String{NewEnumValidator("b64", "gzip")},
			},
			"path": schema.StringAttribute{
				Required:            true,
				Description:         "The path to the file on the target system",
				MarkdownDescription: "The path to the file on the target system",
			},
			"user": schema.StringAttribute{
				Required:            true,
				Description:         "The owning user of the target file",
				MarkdownDescription: "The owning user of the target file",
			},
			"group": schema.StringAttribute{
				Required:    true,
				Description: "The owning group of the target file",
			},
			"permissions": schema.StringAttribute{
				Required:            true,
				Description:         "The permissions of the target file. Must be an octal string between 0 and 77777",
				MarkdownDescription: "The permissions of the target file. Must be an octal string between 0 and 77777",
				Validators:          []validator.String{NewUnixPermissionsValidator()},
			},
			"encode_as": schema.StringAttribute{
				Optional:            true,
				Description:         "If set, it assumes file_contents is valid HCL and will be converted to json or yaml",
				MarkdownDescription: "If set, it assumes file_contents is valid HCL and will be converted to json or yaml",
				Validators:          []validator.String{NewEnumValidator("json", "yaml")},
			},
			"file_contents": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The contents of the file to be written. Omit to write an empty file to the target system",
				MarkdownDescription: "The contents of the file to be written. Omit to write an empty file to the target system",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The SHA-1 checksum of the generated yaml. Used for tracking state",
				MarkdownDescription: "The SHA-1 checksum of the generated yaml. Used for tracking state",
			},
			"content": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "The rendered YAML of the directive, suitable for inserting into a cloud-init's write_files list",
				MarkdownDescription: "The rendered YAML of the directive, suitable for inserting into a cloud-init's write_files list",
			},
			"content_sha256": schema.StringAttribute{
				Computed:            true,
				Description:         "The SHA-256 checksum of the generated yaml",
				MarkdownDescription: "The SHA-256 checksum of the generated yaml",
			},
		},
	}
}

func NewWriteFileResource() resource.Resource {
	return new(WriteFileResource)
}

var _ resource.Resource = (*WriteFileResource)(nil)
