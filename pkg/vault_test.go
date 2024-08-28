package vault

import (
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReplicateKVSecrets(t *testing.T) {
	srcClient := new(MockVaultClient)
	destClient := new(MockVaultClient)

	srcClient.On("Logical").Return(new(MockLogicalClient))
	destClient.On("Logical").Return(new(MockLogicalClient))

	srcClient.Logical().On("List", "secret/metadata/").Return(&api.Secret{
		Data: map[string]interface{}{
			"keys": []interface{}{"myapp/"},
		},
	}, nil)

	srcClient.Logical().On("Read", "secret/data/myapp/").Return(&api.Secret{
		Data: map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
			},
		},
	}, nil)

	destClient.Logical().On("Write", "secret/data/myapp/", map[string]interface{}{
		"data": map[string]interface{}{
			"key": "value",
		},
	}).Return(&api.Secret{}, nil)

	err := ReplicateKVSecrets("srcAddr", "destAddr", "srcToken", "destToken", "myapp/")
	assert.NoError(t, err)

	srcClient.AssertExpectations(t)
	destClient.AssertExpectations(t)
}

func TestCreateVaultClient(t *testing.T) {
	client, err := createVaultClient("http://localhost:8200", "root-token")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8200", client.Address())
}

func TestListSecrets(t *testing.T) {
	client := new(MockVaultClient)
	client.On("Logical").Return(new(MockLogicalClient))

	client.Logical().On("List", "secret/metadata/myapp/").Return(&api.Secret{
		Data: map[string]interface{}{
			"keys": []interface{}{"key1", "key2"},
		},
	}, nil)

	keys, err := listSecrets(client, "myapp/")
	assert.NoError(t, err)
	assert.Equal(t, []string{"key1", "key2"}, keys)

	client.AssertExpectations(t)
}

func TestReadSecret(t *testing.T) {
	client := new(MockVaultClient)
	client.On("Logical").Return(new(MockLogicalClient))

	client.Logical().On("Read", "secret/data/myapp/key1").Return(&api.Secret{
		Data: map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
			},
		},
	}, nil)

	data, err := readSecret(client, "myapp/key1")
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"key": "value"}, data)

	client.AssertExpectations(t)
}

func TestWriteSecret(t *testing.T) {
	client := new(MockVaultClient)
	client.On("Logical").Return(new(MockLogicalClient))

	client.Logical().On("Write", "secret/data/myapp/key1", mock.AnythingOfType("map[string]interface {}")).Return(&api.Secret{}, nil)

	err := writeSecret(client, "myapp/key1", map[string]interface{}{"key": "value"})
	assert.NoError(t, err)

	client.AssertExpectations(t)
}
