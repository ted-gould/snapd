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
	"github.com/snapcore/snapd/snap"
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

	_, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)
}

func (s *Unity8InterfaceSuite) TestGetNameMissing(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
`)

	_, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)
}
func (s *Unity8InterfaceSuite) TestGetNameBadDot(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  name: foo.bar
`)

	_, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)
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

	_, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)
}

func (s *Unity8InterfaceSuite) TestGetNameUnknownAttribute(c *C) {
	var mockSnapYaml = []byte(`name: unity8-client
version: 1.0
slots:
 unity8-slot:
  interface: unity8
  unknown: foo
`)

	_, err := snap.InfoFromSnapYaml(mockSnapYaml)
	c.Assert(err, IsNil)
}

// The label glob when all apps are bound to the unity8 slot
func (s *Unity8InterfaceSuite) TestConnectedPlugSnippetUsesSlotLabelAll(c *C) {
}

// The label uses alternation when some, but not all, apps are bound to the unity8 slot
func (s *Unity8InterfaceSuite) TestConnectedPlugSnippetUsesSlotLabelSome(c *C) {
}

func (s *Unity8InterfaceSuite) TestConnectedPlugSecComp(c *C) {
}

// The label uses short form when exactly one app is bound to the unity8 slot
func (s *Unity8InterfaceSuite) TestConnectedPlugSnippetUsesSlotLabelOne(c *C) {
}

// The label glob when all apps are bound to the unity8 plug
func (s *Unity8InterfaceSuite) TestConnectedSlotSnippetUsesPlugLabelAll(c *C) {
}

// The label uses alternation when some, but not all, apps is bound to the unity8 plug
func (s *Unity8InterfaceSuite) TestConnectedSlotSnippetUsesPlugLabelSome(c *C) {
}

// The label uses short form when exactly one app is bound to the unity8 plug
func (s *Unity8InterfaceSuite) TestConnectedSlotSnippetUsesPlugLabelOne(c *C) {
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmor(c *C) {
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmorWithName(c *C) {
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmorNative(c *C) {
}

func (s *Unity8InterfaceSuite) TestPermanentSlotAppArmorClassic(c *C) {
}

func (s *Unity8InterfaceSuite) TestPermanentSlotSecComp(c *C) {
}

func (s *Unity8InterfaceSuite) TestUnusedSecuritySystems(c *C) {
}

func (s *Unity8InterfaceSuite) TestUsedSecuritySystems(c *C) {
}

func (s *Unity8InterfaceSuite) TestUnexpectedSecuritySystems(c *C) {
}

func (s *Unity8InterfaceSuite) TestAutoConnect(c *C) {
}
