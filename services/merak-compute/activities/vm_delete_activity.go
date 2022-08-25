/*
MIT License
Copyright(c) 2022 Futurewei Cloud

	Permission is hereby granted,
	free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
	to whom the Software is furnished to do so, subject to the following conditions:
	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
package activities

import (
	"context"
	"strconv"
	"strings"

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func VmDelete(ctx context.Context, vmID string) {
	logger := activity.GetLogger(ctx)
	podIP := common.RedisClient.HGet(ctx, vmID, "hostIP").Val()

	var agent_address strings.Builder
	agent_address.Reset()
	agent_address.WriteString(podIP)
	agent_address.WriteString(":")
	agent_address.WriteString(strconv.Itoa(constants.AGENT_GRPC_SERVER_PORT))
	conn, err := grpc.Dial(agent_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Info("Failed to dial gRPC server address: "+agent_address.String(), err)
	}
	client := agent_pb.NewMerakAgentServiceClient(conn)
	port := agent_pb.InternalPortConfig{
		OperationType: common_pb.OperationType_DELETE,
		Name:          common.RedisClient.HGet(ctx, vmID, "name").Val(),
		Projectid:     common.RedisClient.HGet(ctx, vmID, "projectID").Val(),
		Deviceid:      common.RedisClient.HGet(ctx, vmID, "deviceID").Val(),
		Remoteid:      common.RedisClient.HGet(ctx, vmID, "remoteID").Val(),
	}
	logger.Info("Sending to agent: ", podIP)
	resp, err := client.PortHandler(ctx, &port)
	if err != nil {
		logger.Error("Unable delete vm ID " + podIP + "Reason: " + resp.GetReturnMessage() + "\n")
	}
	common.RedisClient.FlushAll(ctx)
}
