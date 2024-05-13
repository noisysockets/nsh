// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";

const wsBaseURL = import.meta.env.PROD
  ? window.location.origin.replace("http", "ws")
  : "ws://localhost:8080";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App wsBaseURL={wsBaseURL} />
  </React.StrictMode>,
);
