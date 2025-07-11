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