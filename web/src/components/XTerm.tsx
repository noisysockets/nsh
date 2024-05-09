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

import React, {
  useEffect,
  useMemo,
  useRef,
  useState,
  CSSProperties,
} from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { EventEmitter } from "eventemitter3";

import "@xterm/xterm/css/xterm.css";
import "./XTerm.css";

interface TerminalDimensions {
  columns: number;
  rows: number;
}

interface XTermProps {
  // input is where the terminal will read data from.
  input: EventEmitter;
  // output is where the terminal will write data to.
  output: EventEmitter;
  // style sets the style of the terminal container.
  style?: CSSProperties;
}

// XTerm is a React component that wraps the xterm.js terminal emulator.
const XTerm: React.FC<XTermProps> = ({ input, output, style }) => {
  const textDecoder = useMemo<TextDecoder>(() => new TextDecoder(), []);
  const textEncoder = useMemo<TextEncoder>(() => new TextEncoder(), []);

  const terminalRef = useRef<HTMLDivElement | null>(null);
  const [term, setTerm] = useState<Terminal | null>(null);
  const fitAddon = useMemo(() => new FitAddon(), []);

  useEffect(() => {
    const terminal = new Terminal({
      cursorBlink: true,
    });

    setTerm(terminal);

    return () => {
      terminal.dispose();
    };
  }, []);

  useEffect(() => {
    if (term) {
      const handleResize = () => {
        fitAddon.fit();
        output.emit("resize", {
          columns: term.cols,
          rows: term.rows,
        } as TerminalDimensions);
      };
      window.addEventListener("resize", handleResize);

      if (terminalRef.current) {
        term.loadAddon(fitAddon);
        term.open(terminalRef.current);
        handleResize();
        term.focus();
        setTerm(term);
      }

      const handleInputData = (data: Uint8Array) => {
        term.write(textDecoder.decode(data));
      };
      input.on("data", handleInputData);

      const handleOutputData = term.onData((data) => {
        output.emit("data", textEncoder.encode(data));
      });

      const handleTitleChange = term.onTitleChange((title: string) => {
        output.emit("title", title);
      });

      return () => {
        window.removeEventListener("resize", handleResize);
        input.off("data", handleInputData);
        handleOutputData.dispose();
        handleTitleChange.dispose();
      };
    }
  }, [term, input, output, textDecoder, textEncoder, fitAddon]);

  return <div ref={terminalRef} style={style} />;
};

export default XTerm;
export type { TerminalDimensions };
