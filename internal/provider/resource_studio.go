package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rodriguezst/terraform-provider-lightningai/internal/client"
)

const (
	defaultStartupTimeout   = "10m"
	pollInterval            = 2 * time.Second
	defaultPollTimeout      = 10 * time.Minute
	stateRunning            = client.StateRunning
	stateStopped            = client.StateStopped
	startupScriptModeOnce   = "once"
	startupScriptModeAlways = "always"
)

var _ resource.Resource = &StudioResource{}

// StudioResource defines the resource implementation.
type StudioResource struct {
	client *client.Client
}

// StudioResourceModel describes the resource data model.
type StudioResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Machine           types.String `tfsdk:"machine"`
	Running           types.Bool   `tfsdk:"running"`
	Interruptible     types.Bool   `tfsdk:"interruptible"`
	StartupScript     types.String `tfsdk:"startup_script"`
	StartupScriptMode types.String `tfsdk:"startup_script_mode"`
	StartupTimeout    types.String `tfsdk:"startup_timeout"`
	Status            types.String `tfsdk:"status"`
	PublicIP          types.String `tfsdk:"public_ip"`
}

// NewStudioResource returns a new studio resource.
func NewStudioResource() resource.Resource {
	return &StudioResource{}
}

func (r *StudioResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_studio"
}

func (r *StudioResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the lifecycle of a Lightning AI Studio.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Studio unique identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Studio name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine": schema.StringAttribute{
				Description: "Machine type used when starting the studio (e.g., cpu-4, lit-l4-1).",
				Optional:    true,
			},
			"running": schema.BoolAttribute{
				Description: "Desired runtime state of the studio. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"interruptible": schema.BoolAttribute{
				Description: "Use spot/preemptible compute.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"startup_script": schema.StringAttribute{
				Description: "Multiline script executed after studio start. Changes trigger resource replacement.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"startup_script_mode": schema.StringAttribute{
				Description: "When to run the startup script: 'once' (only at creation) or 'always' (every start). Defaults to 'once'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(startupScriptModeOnce),
			},
			"startup_timeout": schema.StringAttribute{
				Description: "Maximum time to wait for startup script execution (e.g., 10m, 30m). Defaults to '10m'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(defaultStartupTimeout),
			},
			"status": schema.StringAttribute{
				Description: "Current state of the studio.",
				Computed:    true,
			},
			"public_ip": schema.StringAttribute{
				Description: "Public IP address of the studio, if available.",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

func (r *StudioResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *StudioResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StudioResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Lightning AI Studio", map[string]interface{}{"name": plan.Name.ValueString()})

	studio, err := r.client.CreateStudio(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Studio", err.Error())
		return
	}

	plan.ID = types.StringValue(studio.ID)
	plan.Status = types.StringValue("")

	if plan.Running.ValueBool() {
		machineType := plan.Machine.ValueString()
		interruptible := plan.Interruptible.ValueBool()

		tflog.Debug(ctx, "Starting studio", map[string]interface{}{"id": studio.ID, "machine": machineType})
		if err := r.client.StartStudio(ctx, studio.ID, machineType, interruptible); err != nil {
			resp.Diagnostics.AddError("Error Starting Studio", err.Error())
			return
		}

		if err := r.waitForStatus(ctx, studio.ID, stateRunning, defaultPollTimeout); err != nil {
			resp.Diagnostics.AddError("Error Waiting for Studio to Start", err.Error())
			return
		}

		if !plan.StartupScript.IsNull() && !plan.StartupScript.IsUnknown() && plan.StartupScript.ValueString() != "" {
			if err := r.waitForReady(ctx, studio.ID, defaultPollTimeout); err != nil {
				resp.Diagnostics.AddError("Error Waiting for Studio to Become Ready", err.Error())
				return
			}
			if err := r.executeStartupScript(ctx, studio.ID, plan); err != nil {
				resp.Diagnostics.AddError("Error Executing Startup Script", err.Error())
				return
			}
		}
	}

	if err := r.refreshStatus(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Studio Status", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *StudioResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StudioResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetStudio(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Studio", err.Error())
		return
	}

	if err := r.refreshStatus(ctx, &state); err != nil {
		resp.Diagnostics.AddError("Error Reading Studio Status", err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *StudioResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StudioResourceModel
	var state StudioResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	studioID := state.ID.ValueString()
	plan.ID = state.ID
	plan.Status = state.Status

	currentlyRunning := state.Status.ValueString() == stateRunning

	if plan.Running.ValueBool() && !currentlyRunning {
		machineType := plan.Machine.ValueString()
		interruptible := plan.Interruptible.ValueBool()

		tflog.Debug(ctx, "Starting studio", map[string]interface{}{"id": studioID})
		if err := r.client.StartStudio(ctx, studioID, machineType, interruptible); err != nil {
			resp.Diagnostics.AddError("Error Starting Studio", err.Error())
			return
		}

		if err := r.waitForStatus(ctx, studioID, stateRunning, defaultPollTimeout); err != nil {
			resp.Diagnostics.AddError("Error Waiting for Studio to Start", err.Error())
			return
		}

		if !plan.StartupScript.IsNull() && !plan.StartupScript.IsUnknown() && plan.StartupScript.ValueString() != "" {
			mode := plan.StartupScriptMode.ValueString()
			if mode == startupScriptModeAlways {
				if err := r.waitForReady(ctx, studioID, defaultPollTimeout); err != nil {
					resp.Diagnostics.AddError("Error Waiting for Studio to Become Ready", err.Error())
					return
				}
				if err := r.executeStartupScript(ctx, studioID, plan); err != nil {
					resp.Diagnostics.AddError("Error Executing Startup Script", err.Error())
					return
				}
			}
		}
	} else if !plan.Running.ValueBool() && currentlyRunning {
		tflog.Debug(ctx, "Stopping studio", map[string]interface{}{"id": studioID})
		if err := r.client.StopStudio(ctx, studioID); err != nil {
			resp.Diagnostics.AddError("Error Stopping Studio", err.Error())
			return
		}

		if err := r.waitForStatus(ctx, studioID, stateStopped, defaultPollTimeout); err != nil {
			resp.Diagnostics.AddError("Error Waiting for Studio to Stop", err.Error())
			return
		}
	}

	if err := r.refreshStatus(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Studio Status", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *StudioResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StudioResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	studioID := state.ID.ValueString()
	tflog.Debug(ctx, "Deleting studio", map[string]interface{}{"id": studioID})

	if state.Status.ValueString() == stateRunning {
		if err := r.client.StopStudio(ctx, studioID); err != nil {
			resp.Diagnostics.AddError("Error Stopping Studio Before Delete", err.Error())
			return
		}
		if err := r.waitForStatus(ctx, studioID, stateStopped, defaultPollTimeout); err != nil {
			resp.Diagnostics.AddError("Error Waiting for Studio to Stop Before Delete", err.Error())
			return
		}
	}

	if err := r.client.DeleteStudio(ctx, studioID); err != nil {
		resp.Diagnostics.AddError("Error Deleting Studio", err.Error())
		return
	}
}

// waitForStatus polls until the studio reaches the desired phase or the timeout is reached.
func (r *StudioResource) waitForStatus(ctx context.Context, studioID, desiredStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for studio %s to reach status %s", studioID, desiredStatus)
		}

		status, err := r.client.GetStudioStatus(ctx, studioID)
		if err != nil {
			return fmt.Errorf("error polling studio status: %w", err)
		}

		tflog.Debug(ctx, "Polling studio status", map[string]interface{}{
			"id":     studioID,
			"status": status.Phase,
			"want":   desiredStatus,
		})

		if status.Phase == desiredStatus {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// waitForReady polls until the studio is fully ready (RUNNING + filesystem
// restore complete). This should be called before executing startup scripts
// to ensure persisted user data is available.
func (r *StudioResource) waitForReady(ctx context.Context, studioID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for studio %s to become ready", studioID)
		}

		status, err := r.client.GetStudioStatus(ctx, studioID)
		if err != nil {
			return fmt.Errorf("error polling studio readiness: %w", err)
		}

		tflog.Debug(ctx, "Polling studio readiness", map[string]interface{}{
			"id":                     studioID,
			"phase":                  status.Phase,
			"startupPercentage":      status.StartupPercentage,
			"initialRestoreFinished": status.StartupStatus != nil && status.StartupStatus.InitialRestoreFinished,
		})

		if status.IsReady() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// executeStartupScript wraps and executes the startup script in the studio.
func (r *StudioResource) executeStartupScript(ctx context.Context, studioID string, plan StudioResourceModel) error {
	script := plan.StartupScript.ValueString()
	timeoutStr := plan.StartupTimeout.ValueString()

	scriptTimeout := defaultPollTimeout
	if timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err == nil {
			scriptTimeout = d
		}
	}

	command := fmt.Sprintf("cat <<'EOF' > /tmp/lightning-startup.sh\n%s\nEOF\nbash /tmp/lightning-startup.sh", script)

	execCtx, cancel := context.WithTimeout(ctx, scriptTimeout)
	defer cancel()

	tflog.Debug(ctx, "Executing startup script", map[string]interface{}{"studio_id": studioID})
	result, err := r.client.ExecuteCommand(execCtx, studioID, command)
	if err != nil {
		return fmt.Errorf("startup script execution failed: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("startup script exited with code %d: %s", result.ExitCode, result.Output)
	}

	return nil
}

// refreshStatus updates the status, running, and public IP fields in the model
// based on the actual state from the API, enabling drift detection.
func (r *StudioResource) refreshStatus(ctx context.Context, model *StudioResourceModel) error {
	status, err := r.client.GetStudioStatus(ctx, model.ID.ValueString())
	if err != nil {
		return err
	}

	model.Status = types.StringValue(status.Phase)
	model.Running = types.BoolValue(status.Phase == stateRunning)

	if status.PublicIP != "" {
		model.PublicIP = types.StringValue(status.PublicIP)
	} else {
		model.PublicIP = types.StringNull()
	}

	return nil
}
