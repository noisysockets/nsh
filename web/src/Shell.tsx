// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

import React from "react";
import { EventEmitter } from "eventemitter3";
import { useEffect, useMemo, useState } from "react";
import { Client } from "@noisysockets/shell";

import XTerm, { TerminalDimensions } from "./components/XTerm";

interface ShellProps {
  wsURL: string;
}

const Shell: React.FC<ShellProps> = ({ wsURL }) => {
  const client = useMemo<Client>(() => new Client(wsURL), [wsURL]);
  const [terminalIsOpen, setTerminalIsOpen] = useState(false);
  const [terminalDimensions, setTerminalDimensions] =
    useState<TerminalDimensions>({ columns: 80, rows: 24 });

  const input = useMemo<EventEmitter>(() => new EventEmitter(), []);
  const output = useMemo<EventEmitter>(() => new EventEmitter(), []);

  // Set the title of the document to the title of the terminal.
  useEffect(() => {
    output.on("title", (title: string) => {
      document.title = title;
    });
    return () => {
      output.removeListener("title", (title: string) => {
        document.title = title;
      });
    };
  }, [output]);

  // Set the terminal dimensions when the terminal is resized.
  useEffect(() => {
    output.on("resize", setTerminalDimensions);
    return () => {
      output.removeListener("resize", setTerminalDimensions);
    };
  }, [output, setTerminalDimensions]);

  // Open the terminal when the component is mounted, or resize the terminal
  // if it is already open and the terminal dimensions have changed.
  useEffect(() => {
    if (!terminalIsOpen) {
      // TODO: handle terminal exits, displaying a message to the user.
      client.openTerminal(
        terminalDimensions.columns,
        terminalDimensions.rows,
        ["TERM=xterm-256color"],
        output,
        input,
      );

      setTerminalIsOpen(true);
    } else {
      client.resizeTerminal(
        terminalDimensions.columns,
        terminalDimensions.rows,
      );
    }
  }, [client, terminalIsOpen, terminalDimensions, input, output]);

  return (
    <>
      <XTerm
        input={input}
        output={output}
        style={{ height: "100%", width: "100%", overflow: "hidden" }}
      />
    </>
  );
};

export default Shell;
