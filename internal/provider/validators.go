package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type unixPermissionValidator struct{}

// Description implements [validator.String].
func (u *unixPermissionValidator) Description(_ context.Context) string {
	return "value must be an unsigend 32-bit octal integer between 0 and 77777"
}

// MarkdownDescription implements [validator.String].
func (u *unixPermissionValidator) MarkdownDescription(ctx context.Context) string {
	return u.Description(ctx)
}

// ValidateString implements [validator.String].
func (u *unixPermissionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			u.Description(ctx),
			req.ConfigValue.String(),
		))
		return
	}

	value := req.ConfigValue.ValueString()

	_, err := strconv.ParseUint(value, 8, 32)
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			u.Description(ctx),
			err.Error(),
		))
		return
	}
}

func NewUnixPermissionsValidator() validator.String {
	return new(unixPermissionValidator)
}

type enumValidator struct {
	allowed []string
}

// Description implements [validator.String].
func (e *enumValidator) Description(context.Context) string {
	return fmt.Sprintf("value must be null or one of %s", strings.Join(e.allowed, ", "))
}

// MarkdownDescription implements [validator.String].
func (e *enumValidator) MarkdownDescription(ctx context.Context) string {
	return e.Description(ctx)
}

// ValidateString implements [validator.String].
func (e *enumValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.IsUnknown() {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			e.Description(ctx),
			req.ConfigValue.String(),
		))
		return
	}

	val := req.ConfigValue.ValueString()
	if !slices.Contains(e.allowed, val) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			e.Description(ctx),
			val,
		))
	}
}

func NewEnumValidator(allowed ...string) validator.String {
	return &enumValidator{allowed}
}
