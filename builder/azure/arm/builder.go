package azurearm

import (
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "packer.azure.arm"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ARMConfig           `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"command_wrapper",
				"mount_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Defaults
	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-amazon-chroot-volumes/{{.Device}}"
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.AMIConfig.Prepare(&b.config.ctx)...)

	// Do stuff

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	// Show config without secrets
	log.Println(common.ScrubConfig(b.config, b.config.AccessKey, b.config.SecretKey))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {

	config, err := b.config.Config()
	if err != nil {
		return nil, err
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepPreValidate{
			DestAmiName:     b.config.AMIName,
			ForceDeregister: b.config.AMIForceDeregister,
		},
		&StepInstanceInfo{},
		&awscommon.StepSourceAMIInfo{
			SourceAmi:          b.config.SourceAmi,
			EnhancedNetworking: b.config.AMIEnhancedNetworking,
		},
		&StepCheckRootDevice{},
		&StepFlock{},
		&StepPrepareDevice{},
		&StepCreateVolume{
			RootVolumeSize: b.config.RootVolumeSize,
		},
		&StepAttachVolume{},
		&StepEarlyUnflock{},
		&StepMountDevice{
			MountOptions: b.config.MountOptions,
		},
		&StepMountExtra{},
		&StepCopyFiles{},
		&StepChrootProvision{},
		&StepEarlyCleanup{},
		&StepSnapshot{},
		&awscommon.StepDeregisterAMI{
			ForceDeregister: b.config.AMIForceDeregister,
			AMIName:         b.config.AMIName,
		},
		&StepRegisterAMI{
			RootVolumeSize: b.config.RootVolumeSize,
		},
		&awscommon.StepAMIRegionCopy{
			AccessConfig: &b.config.AccessConfig,
			Regions:      b.config.AMIRegions,
			Name:         b.config.AMIName,
		},
		&awscommon.StepModifyAMIAttributes{
			Description: b.config.AMIDescription,
			Users:       b.config.AMIUsers,
			Groups:      b.config.AMIGroups,
		},
		&awscommon.StepCreateTags{
			Tags: b.config.AMITags,
		},
	}

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no AMIs, then just return
	if _, ok := state.GetOk("amis"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state.Get("amis").(map[string]string),
		BuilderIdValue: BuilderId,
		Conn:           ec2conn,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
