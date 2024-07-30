package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"log"
	"net/http"
	"terraform-provider-opaasn8n/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &workFlowResource{}
	_ resource.ResourceWithConfigure = &workFlowResource{}
)

func NewWorkflowResource() resource.Resource {
	return &workFlowResource{}
}

type workFlowResource struct {
	client *tools.N8NClient
}

type workFlowModel struct {
	ID       types.String `tfsdk:"id"`
	WORKFLOW types.String `tfsdk:"workflow"`
}

type workflowJsonModel struct {
	ID types.String `json:"id"`
}

// Configure adds the provider configured client to the resource.
func (r *workFlowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tools.N8NClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *workFlowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

// Schema defines the schema for the resource.
func (r *workFlowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"workflow": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Create a new resource.
func (r *workFlowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workFlowModel
	jsonModel := &workflowJsonModel{}
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]string{"certificate": plan.WORKFLOW.ValueString()}
	jsonData, _ := json.Marshal(values)

	request, err := http.NewRequest("POST", r.client.Url, bytes.NewBuffer(jsonData))
	request.Header.Set("X-N8N-API-KEY", r.client.Token)
	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		resp.Diagnostics.AddError("Cannot send post request", err.Error())
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	bodyString := string(bodyBytes)
	if response.StatusCode != 200 {
		resp.Diagnostics.AddError("Not created", bodyString)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	_ = json.Unmarshal(bodyBytes, &jsonModel)

	plan.ID = types.StringValue(jsonModel.ID.ValueString())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *workFlowResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workFlowResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workFlowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workFlowModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	request, err := http.NewRequest("DELETE", r.client.Url+"/"+state.ID.ValueString(), nil)
	request.Header.Set("X-N8N-API-KEY", r.client.Token)
	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		resp.Diagnostics.AddError("Cannot send delete request", err.Error())
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	if response.StatusCode != 200 {
		resp.Diagnostics.AddError("Not deleted", bodyString)
	}

	if resp.Diagnostics.HasError() {
		return
	}
}
