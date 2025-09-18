// Kubernetes Web Terminal - Client-side JavaScript
class KubernetesWebTerminal {
    constructor() {
        this.uploadedFiles = new Map(); // Store uploaded files
        this.selectedPod = null;
        this.terminalSocket = null;
        
        this.initializeEventListeners();
        this.loadPods();
    }

    initializeEventListeners() {
        // File upload event listeners
        this.setupDragAndDrop();
        this.setupFileSelection();
        
        // Pod selection event listeners
        document.getElementById('refresh-pods').addEventListener('click', () => this.loadPods());
        
        // Terminal event listeners
        document.getElementById('connect-terminal').addEventListener('click', () => this.connectTerminal());
        document.getElementById('disconnect-terminal').addEventListener('click', () => this.disconnectTerminal());
    }

    setupDragAndDrop() {
        const dropZone = document.getElementById('drop-zone');
        
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
            document.getElementById('file-input').click();
        });
    }

    setupFileSelection() {
        const fileInput = document.getElementById('file-input');
        const fileSelectBtn = document.getElementById('file-select-btn');
        
        fileSelectBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            fileInput.click();
        });
        
        fileInput.addEventListener('change', (e) => {
            this.handleFiles(e.target.files);
        });
    }

    preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    highlight(element) {
        element.classList.add('drag-over');
    }

    unhighlight(element) {
        element.classList.remove('drag-over');
    }

    handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        this.handleFiles(files);
    }

    handleFiles(files) {
        Array.from(files).forEach(file => this.uploadFile(file));
    }

    async uploadFile(file) {
        const fileId = this.generateFileId();
        const fileItem = this.createFileItem(fileId, file);
        
        try {
            // Add file to UI immediately
            this.addFileToList(fileItem);
            
            // Create FormData for upload
            const formData = new FormData();
            formData.append('file', file);
            formData.append('fileId', fileId);
            
            // Upload file to server
            const response = await fetch('/api/upload', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) {
                throw new Error(`Upload failed: ${response.statusText}`);
            }
            
            const result = await response.json();
            
            // Store file information
            this.uploadedFiles.set(fileId, {
                id: fileId,
                name: file.name,
                size: file.size,
                serverPath: result.path,
                uploaded: true
            });
            
            // Update file item status
            this.updateFileItemStatus(fileId, 'success', 'File uploaded successfully');
            
        } catch (error) {
            console.error('Upload error:', error);
            this.updateFileItemStatus(fileId, 'error', error.message);
        }
    }

    generateFileId() {
        return 'file-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    createFileItem(fileId, file) {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.id = fileId;
        
        fileItem.innerHTML = `
            <div class="file-info">
                <div class="file-name">${this.escapeHtml(file.name)}</div>
                <div class="file-size">${this.formatFileSize(file.size)}</div>
                <div class="upload-progress">
                    <div class="upload-progress-bar" style="width: 0%"></div>
                </div>
            </div>
            <div class="file-actions">
                <button class="btn btn-primary btn-small mount-btn" disabled>Mount</button>
                <button class="btn btn-danger btn-small remove-btn">Remove</button>
            </div>
        `;
        
        // Add event listeners
        const mountBtn = fileItem.querySelector('.mount-btn');
        const removeBtn = fileItem.querySelector('.remove-btn');
        
        mountBtn.addEventListener('click', () => this.mountFile(fileId));
        removeBtn.addEventListener('click', () => this.removeFile(fileId));
        
        return fileItem;
    }

    addFileToList(fileItem) {
        const fileList = document.getElementById('file-list');
        fileList.appendChild(fileItem);
    }

    updateFileItemStatus(fileId, status, message) {
        const fileItem = document.getElementById(fileId);
        if (!fileItem) return;
        
        const progressBar = fileItem.querySelector('.upload-progress-bar');
        const mountBtn = fileItem.querySelector('.mount-btn');
        
        if (status === 'success') {
            progressBar.style.width = '100%';
            progressBar.style.backgroundColor = '#27ae60';
            mountBtn.disabled = false;
        } else if (status === 'error') {
            progressBar.style.width = '100%';
            progressBar.style.backgroundColor = '#e74c3c';
            
            // Add error message
            const fileInfo = fileItem.querySelector('.file-info');
            const errorDiv = document.createElement('div');
            errorDiv.className = 'file-error';
            errorDiv.style.color = '#e74c3c';
            errorDiv.style.fontSize = '0.8rem';
            errorDiv.textContent = message;
            fileInfo.appendChild(errorDiv);
        }
    }

    async mountFile(fileId) {
        if (!this.selectedPod) {
            this.showStatusMessage('error', 'Please select a pod first');
            return;
        }
        
        const fileInfo = this.uploadedFiles.get(fileId);
        if (!fileInfo) {
            this.showStatusMessage('error', 'File not found');
            return;
        }
        
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
                    targetPath: `/tmp/${fileInfo.name}`
                })
            });
            
            if (!response.ok) {
                throw new Error(`Mount failed: ${response.statusText}`);
            }
            
            const result = await response.json();
            this.showStatusMessage('success', `File ${fileInfo.name} mounted to ${result.targetPath}`);
            
        } catch (error) {
            console.error('Mount error:', error);
            this.showStatusMessage('error', error.message);
        }
    }

    removeFile(fileId) {
        const fileItem = document.getElementById(fileId);
        if (fileItem) {
            fileItem.remove();
        }
        this.uploadedFiles.delete(fileId);
    }

    async loadPods() {
        const podList = document.getElementById('pod-list');
        podList.innerHTML = '<p class="loading">Loading pods...</p>';
        
        try {
            const response = await fetch('/api/pods');
            if (!response.ok) {
                throw new Error(`Failed to load pods: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.displayPods(data.pods || []);
            
        } catch (error) {
            console.error('Error loading pods:', error);
            podList.innerHTML = '<p class="loading">Error loading pods</p>';
        }
    }

    displayPods(pods) {
        const podList = document.getElementById('pod-list');
        
        if (pods.length === 0) {
            podList.innerHTML = '<p class="loading">No pods found</p>';
            return;
        }
        
        podList.innerHTML = '';
        
        pods.forEach(pod => {
            const podItem = document.createElement('div');
            podItem.className = 'pod-item';
            podItem.innerHTML = `
                <div class="pod-name">${this.escapeHtml(pod.name)}</div>
                <div class="pod-namespace">Namespace: ${this.escapeHtml(pod.namespace)}</div>
            `;
            
            podItem.addEventListener('click', () => this.selectPod(pod, podItem));
            podList.appendChild(podItem);
        });
    }

    selectPod(pod, podElement) {
        // Remove previous selection
        document.querySelectorAll('.pod-item').forEach(item => {
            item.classList.remove('selected');
        });
        
        // Select new pod
        podElement.classList.add('selected');
        this.selectedPod = pod;
        
        // Enable terminal connection
        document.getElementById('connect-terminal').disabled = false;
        
        this.showStatusMessage('info', `Selected pod: ${pod.name} (${pod.namespace})`);
    }

    connectTerminal() {
        // Terminal functionality - placeholder for now
        this.showStatusMessage('info', 'Terminal connection feature coming soon');
    }

    disconnectTerminal() {
        // Terminal disconnect functionality - placeholder for now
        this.showStatusMessage('info', 'Terminal disconnection feature coming soon');
    }

    showStatusMessage(type, message) {
        // Remove existing status messages
        document.querySelectorAll('.status-message').forEach(msg => msg.remove());
        
        const statusDiv = document.createElement('div');
        statusDiv.className = `status-message status-${type}`;
        statusDiv.textContent = message;
        
        // Insert at the top of the first panel
        const firstPanel = document.querySelector('.panel');
        firstPanel.insertBefore(statusDiv, firstPanel.firstChild);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (statusDiv.parentNode) {
                statusDiv.remove();
            }
        }, 5000);
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new KubernetesWebTerminal();
});