// Main preload script that imports and exposes all APIs to the renderer
// This file is loaded by Electron and imports our modular API implementations

import './wallet-api';
import './system-api';

// The individual API files handle exposing themselves to the renderer via contextBridge
// No additional code needed here - just importing them executes the exposure logic