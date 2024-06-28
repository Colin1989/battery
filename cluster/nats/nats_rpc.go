package nats

import "fmt"

func getChannel(serverType, serverID string) string {
	return fmt.Sprintf("pitaya/servers/%s/%s", serverType, serverID)
}
