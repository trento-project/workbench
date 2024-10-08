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

type RunnerBuilderFunction func(arguments OperatorArguments, operationID string) *Runner

// map[operatorName]map[operatorVersion]RunnerBuilderFunction
type OperatorsTree map[string]map[string]RunnerBuilderFunction

func extractVersionAndRunnerName(gathererName string) (string, string, error) {
	parts := strings.Split(gathererName, "@")
	if len(parts) == 1 {
		// no version found, just gatherer name
		return parts[0], "", nil
	}
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"could not extract the runner version from %s, version should follow <operatorName>@<version> syntax",
			gathererName,
		)
	}
	return parts[0], parts[1], nil
}

type Registry struct {
	operators OperatorsTree
}

func (m *Registry) GetOperatorRunnerBuilder(name string) (RunnerBuilderFunction, error) {
	operatorName, version, err := extractVersionAndRunnerName(name)
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
	gatherersList := []string{}

	for operatorName, versions := range m.operators {
		operatorVersions := []string{}
		for v := range versions {
			operatorVersions = append(operatorVersions, v)
		}
		sort.Strings(operatorVersions)
		gatherersList = append(
			gatherersList,
			fmt.Sprintf("%s - %s", operatorName, strings.Join(operatorVersions, "/")),
		)
	}

	return gatherersList
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
