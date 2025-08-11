package operator

import (
	"fmt"
	"sort"
	"strings"
)

type OperatorNotFoundError struct {
	Name string
}

func (e *OperatorNotFoundError) Error() string {
	return fmt.Sprintf("operator %s not found", e.Name)
}

type OperatorBuilder func(operationID string, arguments OperatorArguments) Operator

// map[operatorName]map[operatorVersion]OperatorBuilder
type OperatorBuildersTree map[string]map[string]OperatorBuilder

func extractOperatorNameAndVersion(operatorName string) (string, string, error) {
	parts := strings.Split(operatorName, "@")
	if len(parts) == 1 {
		// no version found, just operator name
		return parts[0], "", nil
	}
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"could not extract the operator version from %s, version should follow <operatorName>@<version> syntax",
			operatorName,
		)
	}
	return parts[0], parts[1], nil
}

type Registry struct {
	operators OperatorBuildersTree
}

func NewRegistry(operators OperatorBuildersTree) *Registry {
	return &Registry{
		operators: operators,
	}
}

func (m *Registry) GetOperatorBuilder(name string) (OperatorBuilder, error) {
	operatorName, version, err := extractOperatorNameAndVersion(name)
	if err != nil {
		return nil, err
	}
	if version == "" {
		latestVersion, err := m.getLatestVersionForOperator(name)
		if err != nil {
			return nil, err
		}
		version = latestVersion
	}

	if g, found := m.operators[operatorName][version]; found {
		return g, nil
	}
	return nil, &OperatorNotFoundError{Name: name}
}

func (m *Registry) AvailableOperators() []string {
	operatorList := []string{}

	for operatorName, versions := range m.operators {
		operatorVersions := []string{}
		for v := range versions {
			operatorVersions = append(operatorVersions, v)
		}
		sort.Strings(operatorVersions)
		operatorList = append(
			operatorList,
			fmt.Sprintf("%s - %s", operatorName, strings.Join(operatorVersions, "/")),
		)
	}

	return operatorList
}

func (m *Registry) getLatestVersionForOperator(name string) (string, error) {
	availableOperators, found := m.operators[name]
	if !found {
		return "", &OperatorNotFoundError{Name: name}
	}
	versions := []string{}
	for v := range availableOperators {
		versions = append(versions, v)
	}

	sort.Strings(versions)

	return versions[len(versions)-1], nil
}

func StandardRegistry(options ...BaseOperatorOption) *Registry {
	return &Registry{
		operators: OperatorBuildersTree{
			ClusterMaintenanceChangeOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewClusterMaintenanceChange(arguments, operationID, OperatorOptions[ClusterMaintenanceChange]{
						BaseOperatorOptions: options,
					})
				},
			},

			CrmClusterStartOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewCrmClusterStart(arguments, operationID, OperatorOptions[CrmClusterStart]{
						BaseOperatorOptions: options,
					})
				},
			},

			CrmClusterStopOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewCrmClusterStop(arguments, operationID, OperatorOptions[CrmClusterStop]{
						BaseOperatorOptions: options,
					})
				},
			},

			HostRebootOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewHostReboot(arguments, operationID, OperatorOptions[HostReboot]{
						BaseOperatorOptions: options,
					})
				},
			},

			SapInstanceStartOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewSAPInstanceStart(arguments, operationID, OperatorOptions[SAPInstanceStart]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapInstanceStopOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewSAPInstanceStop(arguments, operationID, OperatorOptions[SAPInstanceStop]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapSystemStartOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewSAPSystemStart(arguments, operationID, OperatorOptions[SAPSystemStart]{
						BaseOperatorOptions: options,
					})
				},
			},
			SapSystemStopOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewSAPSystemStop(arguments, operationID, OperatorOptions[SAPSystemStop]{
						BaseOperatorOptions: options,
					})
				},
			},
			SaptuneApplySolutionOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewSaptuneApplySolution(arguments, operationID, OperatorOptions[SaptuneApplySolution]{
						BaseOperatorOptions: options,
					})
				},
			},
			SaptuneChangeSolutionOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewSaptuneChangeSolution(arguments, operationID, OperatorOptions[SaptuneChangeSolution]{
						BaseOperatorOptions: options,
					})
				},
			},
			PacemakerEnableOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewServiceEnable(PacemakerEnableOperatorName, arguments, operationID, OperatorOptions[ServiceEnable]{
						BaseOperatorOptions: options,
						OperatorOptions: []Option[ServiceEnable]{
							Option[ServiceEnable](WithServiceToEnable(pacemakerServiceName)),
						},
					})
				},
			},
			PacemakerDisableOperatorName: map[string]OperatorBuilder{
				"v1": func(operationID string, arguments OperatorArguments) Operator {
					return NewServiceDisable(PacemakerDisableOperatorName, arguments, operationID, OperatorOptions[ServiceDisable]{
						BaseOperatorOptions: options,
						OperatorOptions: []Option[ServiceDisable]{
							Option[ServiceDisable](WithServiceToDisable(pacemakerServiceName)),
						},
					})
				},
			},
		},
	}
}
