import React from 'react';
import './App.css';
import { Auth, Detail } from "./Auth.js"
import SplitPane from "react-split-pane";



function App() {
  return (
    <SplitPane split="vertical" minSize={300}>
      <div>
        <Auth />
      </div>
      <div>
        <Detail />
      </div>
    </SplitPane>
  );
}

export default App;
