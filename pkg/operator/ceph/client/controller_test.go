/*
Copyright 2016 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"testing"

	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/rook/rook/pkg/clusterd"
	exectest "github.com/rook/rook/pkg/util/exec/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateClient(t *testing.T) {
	context := &clusterd.Context{Executor: &exectest.MockExecutor{}}

	// must specify caps
	p := cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client1", Namespace: "myns"}}
	err := ValidateClient(context, &p)
	assert.NotNil(t, err)

	// must specify name
	p = cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Namespace: "myns"}}
	err = ValidateClient(context, &p)
	assert.NotNil(t, err)

	// must specify namespace
	p = cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client1"}}
	err = ValidateClient(context, &p)
	assert.NotNil(t, err)

	// must specify all caps
	p = cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client1", Namespace: "myns"}}
	p.Spec.Caps.Osd = "allow *"
	p.Spec.Caps.Mon = "allow *"
	err = ValidateClient(context, &p)
	assert.NotNil(t, err)

	// succeed with caps properly defined
	p = cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client1", Namespace: "myns"}}
	p.Spec.Caps.Osd = "allow *"
	p.Spec.Caps.Mon = "allow *"
	p.Spec.Caps.Mds = "allow *"
	err = ValidateClient(context, &p)
	assert.Nil(t, err)
}

func TestCreateClient(t *testing.T) {
	executor := &exectest.MockExecutor{
		MockExecuteCommandWithOutputFile: func(debug bool, actionName, command, outfile string, args ...string) (string, error) {
			return "", nil
		},
	}
	context := &clusterd.Context{Executor: executor}

	p := &cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client1", Namespace: "myns"},
		Spec: cephv1.ClientSpec{
			Caps: cephv1.ClientCaps{
				Osd: "allow *",
				Mon: "allow *",
				Mds: "allow *",
			},
		},
	}
	// c := &cephv1.ClientController{ObjectMeta: metav1.ObjectMeta{context: context, Namespace: "myns"}}

	p.Spec.Caps.Osd = "allow *"
	p.Spec.Caps.Mon = "allow *"
	p.Spec.Caps.Mds = "allow *"

	exists, err := clientExists(context, p)
	assert.False(t, exists)
	err = createClient(context, p)
	assert.Nil(t, err)

	// fail if caps are empty
	p.Spec.Caps.Osd = ""
	p.Spec.Caps.Mon = ""
	p.Spec.Caps.Mds = ""
	err = createClient(context, p)
	assert.NotNil(t, err)
}

func TestUpdateClient(t *testing.T) {
	// try to set empty caps
	old := cephv1.ClientSpec{
		Caps: cephv1.ClientCaps{
			Osd: "allow *",
			Mon: "allow *",
			Mds: "allow *",
		},
	}
	new := cephv1.ClientSpec{
		Caps: cephv1.ClientCaps{
			Osd: "",
			Mon: "allow *",
			Mds: "allow *",
		},
	}
	changed := clientChanged(old, new)
	assert.False(t, changed)

	// try to create user with caps properly set
	old = cephv1.ClientSpec{
		Caps: cephv1.ClientCaps{
			Osd: "allow *",
			Mon: "allow *",
			Mds: "allow *",
		},
	}
	new = cephv1.ClientSpec{
		Caps: cephv1.ClientCaps{
			Osd: "allow rwx pool=test",
			Mon: "allow rwx pool=test",
			Mds: "allow rwx pool=test",
		},
	}
	changed = clientChanged(old, new)
	assert.True(t, changed)
}

func TestDeleteClient(t *testing.T) {
	executor := &exectest.MockExecutor{
		MockExecuteCommandWithOutputFile: func(debug bool, actionName, command, outfile string, args ...string) (string, error) {
			return "", nil
		},
	}
	context := &clusterd.Context{Executor: executor}

	// delete a client that exists
	p := &cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client1", Namespace: "myns"}}
	exists, err := clientExists(context, p)
	assert.Nil(t, err)
	assert.True(t, exists)
	err = deleteClient(context, p)
	assert.Nil(t, err)

	// succeed even if the client doesn't exist
	p = &cephv1.CephClient{ObjectMeta: metav1.ObjectMeta{Name: "client2", Namespace: "myns"}}
	exists, err = clientExists(context, p)
	assert.Nil(t, err)
	assert.False(t, exists)
	err = deleteClient(context, p)
	assert.Nil(t, err)
}

// func TestGetClientObject(t *testing.T) {
//
// }

func TestGetClientObject(t *testing.T) {
	// get a current version pool object, should return with no error and no migration needed
	pool, err := getClientObject(&cephv1.CephBlockPool{})
	assert.NotNil(t, pool)
	assert.Nil(t, err)

	// try to get an object that isn't a pool, should return with an error
	pool, err = getClientObject(&map[string]string{})
	assert.Nil(t, pool)
	assert.NotNil(t, err)
}
