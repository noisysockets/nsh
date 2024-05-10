/*
 * Copyright 2024 Damian Peckett <damian@pecke.tt>
 *
 * Licensed under the Noisy Sockets Source License 1.0 (NSSL-1.0); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at
 *
 * https://github.com/noisysockets/nsh/blob/main/LICENSE
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

import { EventEmitter } from "eventemitter3";
import { useEffect, useMemo, useState } from "react";
import { Client } from "@noisysockets/shell";

import XTerm, { TerminalDimensions } from "./components/XTerm";

interface AppProps {
  wsURL: string;
}

const App = ({ wsURL }: AppProps): React.ReactElement => {
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

export default App;
