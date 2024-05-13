// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

import React from "react";
import {
  BrowserRouter as Router,
  Navigate,
  Route,
  Routes,
} from "react-router-dom";

import Shell from "./Shell";

interface AppProps {
  wsBaseURL: string;
}

const App = ({ wsBaseURL }: AppProps): React.ReactElement => {
  return (
    <Router>
      <Routes>
        <Route
          path="/shell/"
          element={<Shell wsURL={wsBaseURL + "/shell/ws"} />}
        />
        <Route path="*" element={<Navigate replace to="/shell/" />} />
      </Routes>
    </Router>
  );
};

export default App;
