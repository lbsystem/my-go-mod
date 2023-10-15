package routeutils

import (
	"fmt"
	"log"

	"net"
	"net/netip"
	"time"

	arping "github.com/mdlayher/arp"
	"github.com/moznion/go-iprtb"
	"github.com/moznion/go-optional"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type UserRouteTable struct {
	route_table *iprtb.RouteTable
}

func NewUserRouteTable() UserRouteTable {
	r := UserRouteTable{route_table: iprtb.NewRouteTable()}
	r.init()
	return r
}
func (r *UserRouteTable) List() {
	fmt.Println(r.route_table.DumpRouteTable(nil))
}
func (r *UserRouteTable) route_table_add(destination *net.IPNet,

	gateway net.IP,
	networkInterface string,
	metric int) {
	routingTable := r.route_table
	if destination == nil {
		destination = &net.IPNet{
			IP:   net.ParseIP("0.0.0.0"),
			Mask: net.IPv4Mask(0, 0, 0, 0),
		}

	}
	new_route := &iprtb.Route{Destination: destination, Gateway: gateway,
		Metric: metric, NetworkInterface: networkInterface}
	// 添加路由

	routingTable.AddRoute(nil, new_route)
}

func (r *UserRouteTable) RouteDel(destNet *net.IPNet) (optional.Option[iprtb.Route], error) {
	return r.route_table.RemoveRoute(nil, destNet)

}
func (r *UserRouteTable) Find(target net.IP) (optional.Option[iprtb.Route], error) {
	return r.route_table.MatchRoute(nil, target)
}

func (r *UserRouteTable) init() {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		fmt.Println("Failed to get routes:", err)
		return
	}

	for _, route := range routes {
		ifname, err := net.InterfaceByIndex(route.LinkIndex)
		if err == nil {
			r.route_table_add(route.Dst, route.Gw, ifname.Name, route.Priority)
		} else {
			fmt.Println(err.Error())
		}

	}
	go r.subcribe_route()
}

func (r *UserRouteTable) subcribe_route() {
	// Subscribe to netlink updates
	updates := make(chan netlink.RouteUpdate)
	done := make(chan struct{})
	defer close(done)

	if err := netlink.RouteSubscribe(updates, done); err != nil {
		log.Fatalf("Failed to subscribe to netlink updates: %v", err)
	}

	fmt.Println("Listening for route updates...")
	for update := range updates {
		switch update.Type {
		case unix.RTM_NEWROUTE:
			r.RouteDel(update.Dst)
			fmt.Println("Added a new route:")
		case unix.RTM_DELROUTE:
			r.RouteDel(update.Dst)
			fmt.Println("Deleted a route:")

		default:
			fmt.Println("Unknown route update type:")
		}
	}
}
func GetMACByARP(ipStr string, iface *net.Interface) (net.HardwareAddr, error) {
	ip, err := netip.ParseAddr(ipStr)

	if err != nil {
		return nil, err
	}

	client, err := arping.Dial(iface)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Sending the ARP request
	if err := client.Request(ip); err != nil {
		return nil, err
	}

	// Setting up a timeout for the ARP reply

	// Listening for the ARP reply
	timeout := time.After(35 * time.Millisecond)
	ticker := time.NewTicker(2 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("ARP reply timeout")
		case <-ticker.C:
			packet, _, err := client.Read()
			if err != nil {
				// Handle error here if needed, or just continue
				continue
			}
			if packet.Operation == arping.OperationReply && packet.SenderIP == ip {
				return packet.TargetHardwareAddr, nil
			}
		}
	}
}


	// 添加路由
func RIB_route_add(route *netlink.Route) error{
	
	err := netlink.RouteAdd(route)

	if err != nil {
		fmt.Println("Failed to add route:", err)
		return err
	}
	fmt.Println("Route added successfully")
	return err

	
}

func RIB_route_remove(route *netlink.Route) error{
	err := netlink.RouteDel(route)
	if  err != nil {
		fmt.Println("Failed to delete route:", err)
		return err
	}
	fmt.Println("Route deleted successfully")
	return err
}