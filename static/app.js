// Application state
let currentScript = null;
let savedScripts = [];

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    loadPods();
    loadSavedScripts();
    
    // Set up script editor
    const scriptEditor = document.getElementById('script-editor');
    scriptEditor.addEventListener('input', function() {
        // Auto-save functionality could be added here
    });
});

// Tab functionality
function openTab(evt, tabName) {
    const tabContents = document.getElementsByClassName("tab-content");
    for (let i = 0; i < tabContents.length; i++) {
        tabContents[i].classList.remove("active");
    }
    
    const tabButtons = document.getElementsByClassName("tab-button");
    for (let i = 0; i < tabButtons.length; i++) {
        tabButtons[i].classList.remove("active");
    }
    
    document.getElementById(tabName).classList.add("active");
    evt.currentTarget.classList.add("active");
}

// Terminal functionality
async function loadPods() {
    try {
        const response = await fetch('/api/pods');
        const data = await response.json();
        
        const podSelect = document.getElementById('pod-select');
        podSelect.innerHTML = '';
        
        if (data.pods && data.pods.length > 0) {
            data.pods.forEach(pod => {
                const option = document.createElement('option');
                option.value = pod.name;
                option.textContent = `${pod.name} (${pod.namespace})`;
                podSelect.appendChild(option);
            });
        } else {
            const option = document.createElement('option');
            option.value = '';
            option.textContent = 'No pods found';
            podSelect.appendChild(option);
        }
    } catch (error) {
        console.error('Error loading pods:', error);
        const podSelect = document.getElementById('pod-select');
        podSelect.innerHTML = '<option value="">Error loading pods</option>';
    }
}

function connectToTerminal() {
    const podSelect = document.getElementById('pod-select');
    const selectedPod = podSelect.value;
    
    if (!selectedPod) {
        alert('Please select a pod first');
        return;
    }
    
    // TODO: Implement WebSocket terminal connection
    const terminal = document.getElementById('terminal');
    terminal.innerHTML = `<div style="color: #00ff00;">Connecting to pod: ${selectedPod}...</div>`;
    
    // Placeholder for terminal connection
    setTimeout(() => {
        terminal.innerHTML += `<div style="color: #00ff00;">Connected! Terminal functionality will be implemented with WebSocket connection.</div>`;
    }, 1000);
}

// Tool Builder functionality
function saveScript() {
    const scriptName = document.getElementById('script-name').value.trim();
    const scriptType = document.getElementById('script-type').value;
    const scriptContent = document.getElementById('script-editor').value;
    
    if (!scriptName) {
        alert('Please enter a script name');
        return;
    }
    
    if (!scriptContent.trim()) {
        alert('Please enter script content');
        return;
    }
    
    const script = {
        id: Date.now().toString(),
        name: scriptName,
        type: scriptType,
        content: scriptContent,
        createdAt: new Date().toISOString()
    };
    
    // Check if script with same name exists
    const existingIndex = savedScripts.findIndex(s => s.name === scriptName);
    if (existingIndex >= 0) {
        if (confirm(`Script "${scriptName}" already exists. Do you want to overwrite it?`)) {
            savedScripts[existingIndex] = script;
        } else {
            return;
        }
    } else {
        savedScripts.push(script);
    }
    
    // Save to localStorage
    localStorage.setItem('k8s-scripts', JSON.stringify(savedScripts));
    
    // Refresh the scripts list
    renderScriptsList();
    
    // Show success message
    showOutput(`Script "${scriptName}" saved successfully!`, 'success');
}

function loadSavedScripts() {
    const stored = localStorage.getItem('k8s-scripts');
    if (stored) {
        try {
            savedScripts = JSON.parse(stored);
        } catch (error) {
            console.error('Error loading saved scripts:', error);
            savedScripts = [];
        }
    }
    renderScriptsList();
}

function renderScriptsList() {
    const scriptsList = document.getElementById('scripts-list');
    scriptsList.innerHTML = '';
    
    if (savedScripts.length === 0) {
        scriptsList.innerHTML = '<div style="color: #888; text-align: center; padding: 20px;">No saved scripts yet</div>';
        return;
    }
    
    savedScripts.forEach(script => {
        const scriptItem = document.createElement('div');
        scriptItem.className = 'script-item';
        scriptItem.innerHTML = `
            <h4>${script.name}</h4>
            <div class="script-type">${script.type}</div>
            <div class="script-actions">
                <button onclick="loadScript('${script.id}')">Load</button>
                <button onclick="runScriptById('${script.id}')">Run</button>
                <button class="delete" onclick="deleteScript('${script.id}')">Delete</button>
            </div>
        `;
        scriptsList.appendChild(scriptItem);
    });
}

function loadScript(scriptId) {
    const script = savedScripts.find(s => s.id === scriptId);
    if (script) {
        document.getElementById('script-name').value = script.name;
        document.getElementById('script-type').value = script.type;
        document.getElementById('script-editor').value = script.content;
        currentScript = script;
        
        // Highlight the selected script
        document.querySelectorAll('.script-item').forEach(item => {
            item.classList.remove('active');
        });
        event.target.closest('.script-item').classList.add('active');
        
        showOutput(`Script "${script.name}" loaded into editor`, 'info');
    }
}

function runScript() {
    const scriptContent = document.getElementById('script-editor').value.trim();
    const scriptType = document.getElementById('script-type').value;
    
    if (!scriptContent) {
        alert('Please enter script content to run');
        return;
    }
    
    executeScript(scriptContent, scriptType);
}

function runScriptById(scriptId) {
    const script = savedScripts.find(s => s.id === scriptId);
    if (script) {
        executeScript(script.content, script.type);
    }
}

async function executeScript(scriptContent, scriptType) {
    showOutput('Executing script...', 'info');
    
    try {
        // For now, we'll simulate script execution
        // In a real implementation, this would send the script to the backend
        const response = await fetch('/api/execute-script', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                script: scriptContent,
                type: scriptType
            })
        });
        
        if (response.ok) {
            const result = await response.text();
            showOutput(result, 'success');
        } else {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    } catch (error) {
        // Simulate execution for demonstration
        showOutput(`Simulating ${scriptType} script execution:\n\n${scriptContent}\n\n--- Output ---\nScript execution simulated successfully!\nIn a real environment, this would execute the script in a Kubernetes context.`, 'success');
    }
}

function deleteScript(scriptId) {
    const script = savedScripts.find(s => s.id === scriptId);
    if (script && confirm(`Are you sure you want to delete "${script.name}"?`)) {
        savedScripts = savedScripts.filter(s => s.id !== scriptId);
        localStorage.setItem('k8s-scripts', JSON.stringify(savedScripts));
        renderScriptsList();
        
        // Clear editor if the deleted script was loaded
        if (currentScript && currentScript.id === scriptId) {
            document.getElementById('script-name').value = '';
            document.getElementById('script-editor').value = '';
            currentScript = null;
        }
        
        showOutput(`Script "${script.name}" deleted`, 'info');
    }
}

function showOutput(message, type = 'info') {
    const outputContent = document.getElementById('output-content');
    const timestamp = new Date().toLocaleTimeString();
    const typePrefix = type === 'success' ? '✓' : type === 'error' ? '✗' : 'ℹ';
    
    outputContent.textContent = `[${timestamp}] ${typePrefix} ${message}`;
    outputContent.scrollTop = outputContent.scrollHeight;
}

// Keyboard shortcuts for the script editor
document.addEventListener('keydown', function(event) {
    // Ctrl+S to save script
    if (event.ctrlKey && event.key === 's') {
        event.preventDefault();
        saveScript();
    }
    
    // Ctrl+Enter to run script
    if (event.ctrlKey && event.key === 'Enter') {
        event.preventDefault();
        runScript();
    }
});

// Auto-resize textarea
document.getElementById('script-editor').addEventListener('input', function() {
    this.style.height = 'auto';
    this.style.height = Math.max(300, this.scrollHeight) + 'px';
});

// File Mount functionality
class FileUploadManager {
    constructor() {
        this.uploadedFiles = new Map();
        this.selectedPod = null;
        this.initializeFileUpload();
        this.loadPodsForFileMount();
    }

    initializeFileUpload() {
        this.setupDragAndDrop();
        this.setupFileSelection();
        this.setupPodSelection();
        this.setupTerminalForFileMount();
        
        // Refresh pods button
        const refreshBtn = document.getElementById('refresh-pods');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.loadPodsForFileMount());
        }
    }

    setupDragAndDrop() {
        const dropZone = document.getElementById('drop-zone');
        if (!dropZone) return;

        // Prevent default drag behaviors
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            dropZone.addEventListener(eventName, this.preventDefaults, false);
            document.body.addEventListener(eventName, this.preventDefaults, false);
        });

        // Highlight drop zone when item is dragged over it
        ['dragenter', 'dragover'].forEach(eventName => {
            dropZone.addEventListener(eventName, () => this.highlight(dropZone), false);
        });

        ['dragleave', 'drop'].forEach(eventName => {
            dropZone.addEventListener(eventName, () => this.unhighlight(dropZone), false);
        });

        // Handle dropped files
        dropZone.addEventListener('drop', (e) => this.handleDrop(e), false);
        
        // Handle click to open file selector
        dropZone.addEventListener('click', () => {
            const fileInput = document.getElementById('file-input');
            if (fileInput) fileInput.click();
        });
    }

    setupFileSelection() {
        const fileInput = document.getElementById('file-input');
        const fileSelectBtn = document.getElementById('file-select-btn');
        
        if (fileInput) {
            fileInput.addEventListener('change', (e) => this.handleFiles(e.target.files));
        }
        
        if (fileSelectBtn) {
            fileSelectBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                if (fileInput) fileInput.click();
            });
        }
    }

    setupPodSelection() {
        // Set up pod selection for file mount
        // This will be handled by the loadPodsForFileMount method
    }

    setupTerminalForFileMount() {
        const connectBtn = document.getElementById('connect-terminal');
        const disconnectBtn = document.getElementById('disconnect-terminal');
        
        if (connectBtn) {
            connectBtn.addEventListener('click', () => this.connectTerminal());
        }
        
        if (disconnectBtn) {
            disconnectBtn.addEventListener('click', () => this.disconnectTerminal());
        }
    }

    preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    highlight(element) {
        element.classList.add('dragover');
    }

    unhighlight(element) {
        element.classList.remove('dragover');
    }

    handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        this.handleFiles(files);
    }

    async handleFiles(files) {
        [...files].forEach(file => this.uploadFile(file));
    }

    async uploadFile(file) {
        const formData = new FormData();
        formData.append('file', file);

        try {
            // Show upload progress
            this.addFileToList(file, 'uploading');
            
            const response = await fetch('/api/upload', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`Upload failed: ${response.statusText}`);
            }

            const result = await response.json();
            
            // Store uploaded file info
            this.uploadedFiles.set(result.fileId, {
                ...result,
                originalFile: file
            });

            // Update file list
            this.updateFileInList(file, 'uploaded', result);
            
        } catch (error) {
            console.error('Upload error:', error);
            this.updateFileInList(file, 'error', null, error.message);
        }
    }

    addFileToList(file, status) {
        const fileList = document.getElementById('file-list');
        if (!fileList) return;

        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.id = `file-${file.name}`;
        
        fileItem.innerHTML = `
            <div class="file-info">
                <div class="file-name">${file.name}</div>
                <div class="file-size">${this.formatFileSize(file.size)} - ${status}</div>
            </div>
            <div class="file-actions">
                ${status === 'uploaded' ? '<button class="btn btn-primary btn-sm" onclick="fileManager.mountFile(\'' + file.name + '\')">Mount</button>' : ''}
                <button class="btn btn-danger btn-sm" onclick="fileManager.removeFile('${file.name}')">Remove</button>
            </div>
        `;
        
        fileList.appendChild(fileItem);
    }

    updateFileInList(file, status, result = null, error = null) {
        const fileItem = document.getElementById(`file-${file.name}`);
        if (!fileItem) return;

        const statusText = status === 'uploaded' ? 'Ready to mount' : 
                          status === 'error' ? `Error: ${error}` : status;
        
        const sizeDiv = fileItem.querySelector('.file-size');
        if (sizeDiv) {
            sizeDiv.textContent = `${this.formatFileSize(file.size)} - ${statusText}`;
        }

        // Update actions if uploaded successfully
        if (status === 'uploaded' && result) {
            const actionsDiv = fileItem.querySelector('.file-actions');
            if (actionsDiv) {
                actionsDiv.innerHTML = `
                    <button class="btn btn-primary btn-sm" onclick="fileManager.mountFile('${result.fileId}')">Mount</button>
                    <button class="btn btn-danger btn-sm" onclick="fileManager.removeFile('${file.name}')">Remove</button>
                `;
            }
        }
    }

    async mountFile(fileId) {
        if (!this.selectedPod) {
            alert('Please select a pod first');
            return;
        }

        const targetPath = prompt('Enter target path in pod:', '/tmp/uploaded-file');
        if (!targetPath) return;

        try {
            const response = await fetch('/api/mount', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    fileId: fileId,
                    podName: this.selectedPod.name,
                    namespace: this.selectedPod.namespace,
                    targetPath: targetPath
                })
            });

            if (!response.ok) {
                throw new Error(`Mount failed: ${response.statusText}`);
            }

            const result = await response.json();
            alert(`File mounted successfully at ${result.targetPath}`);
            
        } catch (error) {
            console.error('Mount error:', error);
            alert(`Failed to mount file: ${error.message}`);
        }
    }

    removeFile(fileName) {
        const fileItem = document.getElementById(`file-${fileName}`);
        if (fileItem) {
            fileItem.remove();
        }
        
        // Remove from uploaded files map if exists
        for (const [fileId, fileInfo] of this.uploadedFiles) {
            if (fileInfo.originalFile.name === fileName) {
                this.uploadedFiles.delete(fileId);
                break;
            }
        }
    }

    async loadPodsForFileMount() {
        try {
            const response = await fetch('/api/pods');
            const data = await response.json();
            
            const podList = document.getElementById('pod-list');
            if (!podList) return;
            
            podList.innerHTML = '';
            
            if (data.pods && data.pods.length > 0) {
                data.pods.forEach(pod => {
                    const podItem = document.createElement('div');
                    podItem.className = 'pod-item';
                    podItem.onclick = () => this.selectPod(pod, podItem);
                    
                    podItem.innerHTML = `
                        <div class="pod-info">
                            <h4>${pod.name}</h4>
                            <p>Namespace: ${pod.namespace} | Status: ${pod.status || 'Running'}</p>
                        </div>
                    `;
                    
                    podList.appendChild(podItem);
                });
            } else {
                podList.innerHTML = '<p class="loading">No pods found</p>';
            }
        } catch (error) {
            console.error('Failed to load pods:', error);
            const podList = document.getElementById('pod-list');
            if (podList) {
                podList.innerHTML = '<p class="loading">Error loading pods</p>';
            }
        }
    }

    selectPod(pod, element) {
        // Remove previous selection
        document.querySelectorAll('.pod-item').forEach(item => {
            item.classList.remove('selected');
        });
        
        // Add selection to clicked item
        element.classList.add('selected');
        this.selectedPod = pod;
        
        // Enable terminal connect button
        const connectBtn = document.getElementById('connect-terminal');
        if (connectBtn) {
            connectBtn.disabled = false;
        }
    }

    connectTerminal() {
        if (!this.selectedPod) {
            alert('Please select a pod first');
            return;
        }
        
        // For demo purposes, just show a message
        const terminalContainer = document.querySelector('#file-mount-tab .terminal-container');
        if (terminalContainer) {
            terminalContainer.innerHTML = `
                <div style="color: #00ff00; font-family: monospace; padding: 15px;">
                    Connected to ${this.selectedPod.name} in namespace ${this.selectedPod.namespace}<br/>
                    $ echo "Terminal session active"<br/>
                    Terminal session active<br/>
                    $ ls /tmp<br/>
                    uploaded-file<br/>
                    $ _
                </div>
            `;
        }
        
        // Update button states
        document.getElementById('connect-terminal').disabled = true;
        document.getElementById('disconnect-terminal').disabled = false;
    }

    disconnectTerminal() {
        const terminalContainer = document.querySelector('#file-mount-tab .terminal-container');
        if (terminalContainer) {
            terminalContainer.innerHTML = `
                <div class="terminal-placeholder">
                    <p>Select a pod to start terminal session</p>
                </div>
            `;
        }
        
        // Update button states
        document.getElementById('connect-terminal').disabled = false;
        document.getElementById('disconnect-terminal').disabled = true;
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// Initialize file upload manager when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Only initialize if file mount elements exist
    if (document.getElementById('drop-zone')) {
        window.fileManager = new FileUploadManager();
    }
});