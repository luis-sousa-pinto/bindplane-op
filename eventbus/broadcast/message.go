// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package broadcast

// MessageAttributes represents the attributes of a pubsub message.
type MessageAttributes map[string]string

const (
	// AttributeOrigin is the name of the attribute which is set to the node ID of the node which published the message.
	AttributeOrigin = "_origin"
	// AttributeType is the name of the attribute which is set to the type of the message.
	AttributeType = "_type"
	// AttributeAccountID is the name of the attribute which is set to the account ID of the account which published the message.
	AttributeAccountID = "_account_id"
	// AttributeRoutingKey is the name of the attribute which is set to the routing key of the message.
	AttributeRoutingKey = "_routing_key"
)

// Type returns the type of the message.
func (m MessageAttributes) Type() string {
	return m[AttributeType]
}

// ----------------------------------------------------------------------

// Processor is used to process messages before sending to pub/sub and after receiving from pub/sub.
type Processor[T any] interface {
	// AddAttributes is used to add attributes to the message before sending to pub/sub.
	AddAttributes(m *T, attributes MessageAttributes)

	// AcceptMessage is used to determine if the message should be accepted or ignored after receiving from pub/sub.
	AcceptMessage(attributes MessageAttributes) bool

	// OrderingKey specifies the ordering key to use for the message. It should be unique for each message type and must
	// be less than 1KB with the messageType: prefix.
	OrderingKey(m *T) string
}
