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

package builtin_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/builtin"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/testutil"
)

type Unity8InterfaceSuite struct {
	iface interfaces.Interface
	slot  *interfaces.Slot
	plug  *interfaces.Plug
}

var _ = Suite(&Unity8InterfaceSuite{
	iface: &builtin.Unity8Interface{},
	slot: &interfaces.Slot{
		SlotInfo: &snap.SlotInfo{
			Snap:      &snap.Info{SuggestedName: "unity8"},
			Name:      "unity8-session",
			Interface: "unity8",
		},
	},
	plug: &interfaces.Plug{
		PlugInfo: &snap.PlugInfo{
			Snap:      &snap.Info{SuggestedName: "unity8"},
			Name:      "unity8-app",
			Interface: "unity8",
		},
	},
})

func (s *Unity8InterfaceSuite) TestName(c *C) {
	c.Assert(s.iface.Name(), Equals, "unity8")
}

func (s *Unity8InterfaceSuite) TestGetName(c *C) {
	var mockSnapYaml = []byte(`name: unity8-app
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  name: foo
`)

	info, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)

	slot := &interfaces.Slot{SlotInfo: info.Slots["unity8-slot"]}
	iface := &builtin.Unity8Interface{}
}

func (s *Unity8InterfaceSuite) TestGetNameMissing(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
`)

	info, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)

	slot := &interfaces.Slot{SlotInfo: info.Slots["unity8-slot"]}
	iface := &builtin.Unity8Interface{}
	name, err := builtin.Unity8GetName(iface, slot.Attrs)
	c.Assert(err, IsNil)
	c.Assert(name, Equals, "@{SNAP_NAME}")
}
func (s *Unity8InterfaceSuite) TestGetNameBadDot(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  name: foo.bar
`)

	info, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)

	slot := &interfaces.Slot{SlotInfo: info.Slots["unity8-slot"]}
	iface := &builtin.Unity8Interface{}
	name, err := builtin.Unity8GetName(iface, slot.Attrs)
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, "invalid name element: \"foo.bar\"")
	c.Assert(name, Equals, "")
}

func (s *Unity8InterfaceSuite) TestGetNameBadList(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  name:
  - foo
`)

	info, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)

	slot := &interfaces.Slot{SlotInfo: info.Slots["unity8-slot"]}
	iface := &builtin.Unity8Interface{}
	name, err := builtin.Unity8GetName(iface, slot.Attrs)
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, `name element \[foo\] is not a string`)
	c.Assert(name, Equals, "")
}

func (s *Unity8InterfaceSuite) TestGetNameUnknownAttribute(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  unknown: foo
`)

	info, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)

	slot := &interfaces.Slot{SlotInfo: info.Slots["unity8-slot"]}
	iface := &builtin.Unity8Interface{}
	name, err := builtin.Unity8GetName(iface, slot.Attrs)
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, "unknown attribute 'unknown'")
	c.Assert(name, Equals, "")
}

// The label glob when all apps are bound to the unity8 slot
func (s *Unity8InterfaceSuite) TestConnectedPlugSnippetUsesSlotLabelAll(c *C) {
	app1 := &snap.AppInfo{Name: "app1"}
	app2 := &snap.AppInfo{Name: "app2"}
	slot := &interfaces.Slot{
		SlotInfo: &snap.SlotInfo{
			Snap: &snap.Info{
				SuggestedName: "unity8",
				Apps:          map[string]*snap.AppInfo{"app1": app1, "app2": app2},
			},
			Name:      "unity8",
			Interface: "unity8",
			Apps:      map[string]*snap.AppInfo{"app1": app1, "app2": app2},
		},
	}
	snippet, err := s.iface.ConnectedPlugSnippet(s.plug, slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(string(snippet), testutil.Contains, `peer=(label="snap.unity8.*"),`)
}

// The label uses alternation when some, but not all, apps are bound to the unity8 slot
func (s *Unity8InterfaceSuite) TestConnectedPlugSnippetUsesSlotLabelSome(c *C) {
	app1 := &snap.AppInfo{Name: "app1"}
	app2 := &snap.AppInfo{Name: "app2"}
	app3 := &snap.AppInfo{Name: "app3"}
	slot := &interfaces.Slot{
		SlotInfo: &snap.SlotInfo{
			Snap: &snap.Info{
				SuggestedName: "unity8",
				Apps:          map[string]*snap.AppInfo{"app1": app1, "app2": app2, "app3": app3},
			},
			Name:      "unity8",
			Interface: "unity8",
			Apps:      map[string]*snap.AppInfo{"app1": app1, "app2": app2},
		},
	}
	snippet, err := s.iface.ConnectedPlugSnippet(s.plug, slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(string(snippet), testutil.Contains, `peer=(label="snap.unity8.{app1,app2}"),`)
}

func (s *Unity8InterfaceSuite) TestConnectedPlugSecComp(c *C) {
	snippet, err := s.iface.ConnectedPlugSnippet(s.plug, s.slot, interfaces.SecuritySecComp)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))

	c.Check(string(snippet), testutil.Contains, "getsockname\n")
}

// The label uses short form when exactly one app is bound to the unity8 slot
func (s *Unity8InterfaceSuite) TestConnectedPlugSnippetUsesSlotLabelOne(c *C) {
	app := &snap.AppInfo{Name: "app"}
	slot := &interfaces.Slot{
		SlotInfo: &snap.SlotInfo{
			Snap: &snap.Info{
				SuggestedName: "unity8",
				Apps:          map[string]*snap.AppInfo{"app": app},
			},
			Name:      "unity8",
			Interface: "unity8",
			Apps:      map[string]*snap.AppInfo{"app": app},
		},
	}
	snippet, err := s.iface.ConnectedPlugSnippet(s.plug, slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(string(snippet), testutil.Contains, `peer=(label="snap.unity8.app"),`)
}

// The label glob when all apps are bound to the unity8 plug
func (s *Unity8InterfaceSuite) TestConnectedSlotSnippetUsesPlugLabelAll(c *C) {
	app1 := &snap.AppInfo{Name: "app1"}
	app2 := &snap.AppInfo{Name: "app2"}
	plug := &interfaces.Plug{
		PlugInfo: &snap.PlugInfo{
			Snap: &snap.Info{
				SuggestedName: "unity8",
				Apps:          map[string]*snap.AppInfo{"app1": app1, "app2": app2},
			},
			Name:      "unity8",
			Interface: "unity8",
			Apps:      map[string]*snap.AppInfo{"app1": app1, "app2": app2},
		},
	}
	snippet, err := s.iface.ConnectedSlotSnippet(plug, s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(string(snippet), testutil.Contains, `peer=(label="snap.unity8.*"),`)
}

// The label uses alternation when some, but not all, apps is bound to the unity8 plug
func (s *Unity8InterfaceSuite) TestConnectedSlotSnippetUsesPlugLabelSome(c *C) {
	app1 := &snap.AppInfo{Name: "app1"}
	app2 := &snap.AppInfo{Name: "app2"}
	app3 := &snap.AppInfo{Name: "app3"}
	plug := &interfaces.Plug{
		PlugInfo: &snap.PlugInfo{
			Snap: &snap.Info{
				SuggestedName: "unity8",
				Apps:          map[string]*snap.AppInfo{"app1": app1, "app2": app2, "app3": app3},
			},
			Name:      "unity8",
			Interface: "unity8",
			Apps:      map[string]*snap.AppInfo{"app1": app1, "app2": app2},
		},
	}
	snippet, err := s.iface.ConnectedSlotSnippet(plug, s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(string(snippet), testutil.Contains, `peer=(label="snap.unity8.{app1,app2}"),`)
}

// The label uses short form when exactly one app is bound to the unity8 plug
func (s *Unity8InterfaceSuite) TestConnectedSlotSnippetUsesPlugLabelOne(c *C) {
	app := &snap.AppInfo{Name: "app"}
	plug := &interfaces.Plug{
		PlugInfo: &snap.PlugInfo{
			Snap: &snap.Info{
				SuggestedName: "unity8",
				Apps:          map[string]*snap.AppInfo{"app": app},
			},
			Name:      "unity8",
			Interface: "unity8",
			Apps:      map[string]*snap.AppInfo{"app": app},
		},
	}
	snippet, err := s.iface.ConnectedSlotSnippet(plug, s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(string(snippet), testutil.Contains, `peer=(label="snap.unity8.app"),`)
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmor(c *C) {
	snippet, err := s.iface.PermanentSlotSnippet(s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))

	// verify bind rule
	c.Check(string(snippet), testutil.Contains, "dbus (bind)\n    bus=session\n    name=\"org.unity8.MediaPlayer2.@{SNAP_NAME}{,.*}\",\n")
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmorWithName(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  name: foo
`)

	info, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)

	slot := &interfaces.Slot{SlotInfo: info.Slots["unity8-slot"]}
	iface := &builtin.Unity8Interface{}
	snippet, err := iface.PermanentSlotSnippet(slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))

	// verify bind rule
	c.Check(string(snippet), testutil.Contains, "dbus (bind)\n    bus=session\n    name=\"org.unity8.MediaPlayer2.foo{,.*}\",\n")
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmorNative(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()
	iface := &builtin.Unity8Interface{}
	snippet, err := iface.PermanentSlotSnippet(s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))

	// verify classic rule not present
	c.Check(string(snippet), Not(testutil.Contains), "# Allow unconfined clients to interact with the player on classic\n")
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmorClassic(c *C) {
	restore := release.MockOnClassic(true)
	defer restore()
	iface := &builtin.Unity8Interface{}
	snippet, err := iface.PermanentSlotSnippet(s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))

	// verify classic rule present
	c.Check(string(snippet), testutil.Contains, "# Allow unconfined clients to interact with the player on classic\n")
}

func (s *Unity8InterfaceSuite) TestPermanentSlotSecComp(c *C) {
	snippet, err := s.iface.PermanentSlotSnippet(s.slot, interfaces.SecuritySecComp)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))

	c.Check(string(snippet), testutil.Contains, "getsockname\n")
}

func (s *Unity8InterfaceSuite) TestUnusedSecuritySystems(c *C) {
	systems := [...]interfaces.SecuritySystem{interfaces.SecurityDBus,
		interfaces.SecurityUDev, interfaces.SecurityMount}
	for _, system := range systems {
		snippet, err := s.iface.PermanentPlugSnippet(s.plug, system)
		c.Assert(err, IsNil)
		c.Assert(snippet, IsNil)

		snippet, err = s.iface.ConnectedPlugSnippet(s.plug, s.slot, system)
		c.Assert(err, IsNil)
		c.Assert(snippet, IsNil)

		snippet, err = s.iface.PermanentSlotSnippet(s.slot, system)
		c.Assert(err, IsNil)
		c.Assert(snippet, IsNil)

		snippet, err = s.iface.ConnectedSlotSnippet(s.plug, s.slot, system)
		c.Assert(err, IsNil)
		c.Assert(snippet, IsNil)
	}

	snippet, err := s.iface.PermanentPlugSnippet(s.plug, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(snippet, IsNil)

	snippet, err = s.iface.PermanentPlugSnippet(s.plug, interfaces.SecuritySecComp)
	c.Assert(err, IsNil)
	c.Assert(snippet, IsNil)

	snippet, err = s.iface.ConnectedSlotSnippet(s.plug, s.slot, interfaces.SecuritySecComp)
	c.Assert(err, IsNil)
	c.Assert(snippet, IsNil)
}

func (s *Unity8InterfaceSuite) TestUsedSecuritySystems(c *C) {
	systems := [...]interfaces.SecuritySystem{interfaces.SecurityAppArmor,
		interfaces.SecuritySecComp}
	for _, system := range systems {
		snippet, err := s.iface.ConnectedPlugSnippet(s.plug, s.slot, system)
		c.Assert(err, IsNil)
		c.Assert(snippet, Not(IsNil))
		snippet, err = s.iface.PermanentSlotSnippet(s.slot, system)
		c.Assert(err, IsNil)
		c.Assert(snippet, Not(IsNil))
	}
	snippet, err := s.iface.ConnectedSlotSnippet(s.plug, s.slot, interfaces.SecurityAppArmor)
	c.Assert(err, IsNil)
	c.Assert(snippet, Not(IsNil))
}

func (s *Unity8InterfaceSuite) TestUnexpectedSecuritySystems(c *C) {
	snippet, err := s.iface.PermanentPlugSnippet(s.plug, "foo")
	c.Assert(err, Equals, interfaces.ErrUnknownSecurity)
	c.Assert(snippet, IsNil)
	snippet, err = s.iface.ConnectedPlugSnippet(s.plug, s.slot, "foo")
	c.Assert(err, Equals, interfaces.ErrUnknownSecurity)
	c.Assert(snippet, IsNil)
	snippet, err = s.iface.PermanentSlotSnippet(s.slot, "foo")
	c.Assert(err, Equals, interfaces.ErrUnknownSecurity)
	c.Assert(snippet, IsNil)
	snippet, err = s.iface.ConnectedSlotSnippet(s.plug, s.slot, "foo")
	c.Assert(err, Equals, interfaces.ErrUnknownSecurity)
	c.Assert(snippet, IsNil)
}

func (s *Unity8InterfaceSuite) TestAutoConnect(c *C) {
	iface := &builtin.Unity8Interface{}
	c.Check(iface.AutoConnect(), Equals, false)
}
