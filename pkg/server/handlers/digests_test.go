/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of Dnote.
 *
 * Dnote is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Dnote is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Dnote.  If not, see <https://www.gnu.org/licenses/>.
 */

package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/dnote/dnote/pkg/assert"
	"github.com/dnote/dnote/pkg/server/database"
	"github.com/dnote/dnote/pkg/server/testutils"
)

func TestGetDigest(t *testing.T) {
	defer testutils.ClearData()

	// Setup
	server := MustNewServer(t, nil)
	defer server.Close()

	owner := testutils.SetupUserData()
	nonOwner := testutils.SetupUserData()
	digest := database.Digest{
		UserID: owner.ID,
	}
	testutils.MustExec(t, testutils.DB.Save(&digest), "preparing digest")

	t.Run("owner", func(t *testing.T) {
		// Execute
		req := testutils.MakeReq(server.URL, "GET", fmt.Sprintf("/digests/%s", digest.UUID), "")
		res := testutils.HTTPAuthDo(t, req, owner)

		// Test
		assert.StatusCodeEquals(t, res, http.StatusOK, "")
	})

	t.Run("non owner", func(t *testing.T) {
		// Execute
		req := testutils.MakeReq(server.URL, "GET", fmt.Sprintf("/digests/%s", digest.UUID), "")
		res := testutils.HTTPAuthDo(t, req, nonOwner)

		// Test
		assert.StatusCodeEquals(t, res, http.StatusNotFound, "")
	})

	t.Run("guest", func(t *testing.T) {
		// Execute
		req := testutils.MakeReq(server.URL, "GET", fmt.Sprintf("/digests/%s", digest.UUID), "")
		res := testutils.HTTPDo(t, req)

		// Test
		assert.StatusCodeEquals(t, res, http.StatusUnauthorized, "")
	})
}

func TestGetDigests(t *testing.T) {
	defer testutils.ClearData()

	// Setup
	server := MustNewServer(t, nil)
	defer server.Close()

	user := testutils.SetupUserData()
	digest := database.Digest{
		UserID: user.ID,
	}
	testutils.MustExec(t, testutils.DB.Save(&digest), "preparing digest")

	t.Run("user", func(t *testing.T) {
		// Execute
		req := testutils.MakeReq(server.URL, "GET", "/digests", "")
		res := testutils.HTTPAuthDo(t, req, user)

		// Test
		assert.StatusCodeEquals(t, res, http.StatusOK, "")
	})

	t.Run("guest", func(t *testing.T) {
		// Execute
		req := testutils.MakeReq(server.URL, "GET", fmt.Sprintf("/digests/%s", digest.UUID), "")
		res := testutils.HTTPDo(t, req)

		// Test
		assert.StatusCodeEquals(t, res, http.StatusUnauthorized, "")
	})
}