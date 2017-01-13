// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/snap"
)

var unity8ConnectedPlugAppArmor = []byte(`
  #include <abstractions/base>
  #include <abstractions/fonts>
  #include <abstractions/X>

  # Apps fail to start when linked against newer curl/gnutls if we don't allow
  # this. (LP: #1350152)
  #include <abstractions/openssl>

  # Mir-specific stuff
  #include <abstractions/mir>

  # Needed by native GL applications on Mir
  owner /{,var/}run/user/*/mir_socket rw,

  # Hardware-specific accesses
  #include "/usr/share/apparmor/hardware/graphics.d"

  #
  # IPC rules common for all apps
  #
  # Allow connecting to session bus and where to connect to services
  #include <abstractions/dbus-session-strict>

  # Allow connecting to system bus and where to connect to services. Put these
  # here so we don't need to repeat these rules in multiple places (actual
  # communications with any system services is mediated elsewhere). This does
  # allow apps to brute-force enumerate system services, but our system
  # services aren't a secret.
  #include <abstractions/dbus-strict>

  # Unity shell
  dbus (send)
       bus=session
       path="/BottomBarVisibilityCommunicator"
       interface="org.freedesktop.DBus.{Introspectable,Properties}"
       peer=(name=com.canonical.Shell.BottomBarVisibilityCommunicator,label=unconfined),
  dbus (receive)
       bus=session
       path="/BottomBarVisibilityCommunicator"
       interface="com.canonical.Shell.BottomBarVisibilityCommunicator"
       peer=(label=unconfined),
  # on screen keyboard (OSK)
  dbus (send)
       bus=session
       path="/org/maliit/server/address"
       interface="org.freedesktop.DBus.Properties"
       member=Get
       peer=(name=org.maliit.server,label=unconfined),
  unix (connect, receive, send)
       type=stream
       peer=(addr="@/tmp/maliit-server/dbus-*"),

  # clipboard (LP: #1371170)
  dbus (receive, send)
       bus=session
       path="/com/canonical/QtMir/Clipboard"
       interface="com.canonical.QtMir.Clipboard"
       peer=(label=unconfined),
  dbus (receive, send)
       bus=session
       path="/com/canonical/QtMir/Clipboard"
       interface="org.freedesktop.DBus.{Introspectable,Properties}"
       peer=(label=unconfined),

  # usensors
  dbus (send)
       bus=session
       path=/com/canonical/usensord/haptic
       interface=com.canonical.usensord.haptic
       peer=(label=unconfined),

  # URL dispatcher. All apps can call this since:
  # a) the dispatched application is launched out of process and not
  #    controllable except via the specified URL
  # b) the list of url types is strictly controlled
  # c) the dispatched application will launch in the foreground over the
  #    confined app
  dbus (send)
       bus=session
       path="/com/canonical/URLDispatcher"
       interface="com.canonical.URLDispatcher"
       member="DispatchURL"
       peer=(label=unconfined),

  # This is needed when the app is already running and needs to be passed in
  # a URL to open. This is most often used with content-hub providers and
  # url-dispatcher, but is actually supported by Qt generally (though because
  # we don't allow the send a malicious app can't send this to another app).
  dbus (receive)
       bus=session
       path=/###APP_ID_DBUS###
       interface="org.freedesktop.Application"
       member="Open"
       peer=(label=unconfined),

  # This is needed for apps to interact with the Launcher (eg, for the counter)
  dbus (receive, send)
       bus=session
       path=/com/canonical/unity/launcher/###APP_ID_DBUS###
       peer=(label=unconfined),

  # Untrusted Helpers are 3rd party apps that run in a different confinement
  # context and are in a separate Mir session from the calling app (eg, an
  # app that uses a content provider from another app). These helpers use
  # Trusted Prompt Sessions to overlay their window over the calling app and
  # need to get the Mir socket that was setup by the associated trusted helper
  # (eg, content-hub). Typical consumers are content-hub providers,
  # pay-service, url-dispatcher and possibly online-accounts.
  # LP: #1462492 - this rule is suboptimal and should not be needed once we
  # move to socket activation or FD passing
  dbus (receive, send)
       path=/com/canonical/UbuntuAppLaunch/###APP_ID_DBUS###/*
       interface="com.canonical.UbuntuAppLaunch.SocketDemangler"
       member="GetMirSocket"
       bus=session
       peer=(label=unconfined),
  # Allow access to the socket-demangler (needed for the above)
  /usr/lib/@{multiarch}/ubuntu-app-launch/socket-demangler rmix,

  # TODO: finetune this
  dbus (send)
       bus=session
       peer=(name=org.a11y.Bus,label=unconfined),
  dbus (receive)
       bus=session
       interface=org.a11y.atspi**
       peer=(label=unconfined),
  dbus (receive, send)
       bus=accessibility
       peer=(label=unconfined),

  # Deny potentially dangerous access
  deny dbus bus=session
            path=/com/canonical/[Uu]nity/[Dd]ebug**,
  audit deny dbus bus=session
                  interface="com.canonical.snapdecisions",
  deny dbus (send)
       bus=session
       interface="org.gnome.GConf.Server",

  # LP: #1433590
  deny dbus bus=system
            path="/org/freedesktop/Accounts",

  # LP: #1378823
  deny dbus (bind)
       name="org.freedesktop.Application",

  # Allow access to the PasteBoard
  dbus (receive, send)
       bus=session
       interface="com.ubuntu.content.dbus.Service"
       path="/"
       member="CreatePaste"
       peer=(label=unconfined),
  dbus (receive, send)
       bus=session
       interface="com.ubuntu.content.dbus.Service"
       path="/"
       member="GetPasteData"
       peer=(label=unconfined),
  dbus (receive, send)
       bus=session
       interface="com.ubuntu.content.dbus.Service"
       path="/"
       member="GetLatestPasteData"
       peer=(label=unconfined),
  dbus (receive, send)
       bus=session
       interface="com.ubuntu.content.dbus.Service"
       path="/"
       member="PasteFormats"
       peer=(label=unconfined),
  dbus (receive)
       bus=session
       interface="com.ubuntu.content.dbus.Service"
       path="/"
       member="PasteFormatsChanged"
       peer=(label=unconfined),

  #
  # end DBus rules common for all apps
  #
`)

var secCompDBus = []byte(`
# dbus
connect
getsockname
recvmsg
send
sendto
sendmsg
socket
`)

var unity8ConnectedSlotAppArmor = []byte(`
dbus (receive, send)
    bus=session
    peer=(label=###PLUG_SECURITY_TAGS###),
`)

var unity8SlotAppArmor = []byte(`
  # Mir-specific stuff
  #include <abstractions/mir>

  # Needed by native GL applications on Mir
  owner /{,var/}run/user/*/mir_socket rw,
`)

type Unity8Interface struct{}

func (iface *Unity8Interface) Name() string {
	return "unity8"
}

func (iface *Unity8Interface) PermanentPlugSnippet(plug *interfaces.Plug, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	return nil, nil
}

func (iface *Unity8Interface) dbusAppId(info *snap.Info, appname string) []byte {
	var retval io.Writer

	appidbits := []string{info.SideInfo.RealName, appname, info.SideInfo.Revision.String()}
	appid := []byte(strings.Join(appidbits, "_"))

	for value := range appid {
		if value >= 'a' || value <= 'z' || value >= 'A' || value <= 'Z' {
			io.WriteString(retval, string(value))
		} else {
			io.WriteString(retval, fmt.Sprintf("_%2.2d", value))
		}
	}

	return []byte(fmt.Sprint(retval))
}

func (iface *Unity8Interface) ConnectedPlugSnippet(plug *interfaces.Plug, slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor:
		var apparmors []byte
		for _, app := range plug.Apps {
			tag := []byte("###APP_ID_DBUS###")
			value := iface.dbusAppId(plug.PlugInfo.Snap, app.Name)
			apparmors = append(apparmors, bytes.Replace(unity8ConnectedPlugAppArmor, tag, value, -1)...)
		}
		return apparmors, nil
	case interfaces.SecuritySecComp:
		return secCompDBus, nil
	}
	return nil, nil
}

func (iface *Unity8Interface) PermanentSlotSnippet(slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecuritySecComp:
		return secCompDBus, nil
	}
	return nil, nil
}

func (iface *Unity8Interface) ConnectedSlotSnippet(plug *interfaces.Plug, slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor:
		tag := []byte("###PLUG_SECURITY_TAG###")
		value := plugAppLabelExpr(plug)
		return bytes.Replace(unity8SlotAppArmor, tag, value, -1), nil
	}
	return nil, nil
}

func (iface *Unity8Interface) SanitizePlug(plug *interfaces.Plug) error {
	/* Everyone has this, me too! */
	if iface.Name() != plug.Interface {
		panic(fmt.Sprintf("slot is not of interface %q", iface))
	}

	if plug.Snap.Type != "app" {
		return fmt.Errorf("'unity8' plug must be on an application")
	}

	/* Make sure this is tied to applications not packages */
	if len(plug.Apps) == 0 {
		return fmt.Errorf("'unity8' plug must be on an application")
	}

	mountdir := plug.Snap.MountDir()
	for _, app := range plug.Apps {
		/* Check to ensure we have a desktop file for each application */
		path := filepath.Join(mountdir, "meta", "gui", fmt.Sprintf("%s.desktop", app.Name))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("Application '%s' does not have a required desktop file for interface '%s'", app.Name, iface.Name())
		}

		/* Ensure that we're not daemons */
		if app.Daemon != "" {
			return fmt.Errorf("Application '%s' is a daemon, which isn't allowed to have a 'unity8' interface", app.Name)
		}
	}

	return nil
}

func (iface *Unity8Interface) SanitizeSlot(slot *interfaces.Slot) error {
	if iface.Name() != slot.Interface {
		panic(fmt.Sprintf("slot is not of interface %q", iface))
	}

	if slot.Snap.Type != "app" {
		return fmt.Errorf("'unity8' slot must be on an application")
	}

	return nil
}

func (iface *Unity8Interface) AutoConnect(*interfaces.Plug, *interfaces.Slot) bool {
	return true
}
