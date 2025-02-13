function render(request)
    local nodePath = request.params["path"] or ""
    local content = ""
    local error = ""
    
    if nodePath ~= "" then
        content, error = readNode(nodePath)
        if error then
            content = ""
        end
    end
    
    -- Escape special HTML characters
    local function escapeHTML(s)
        return (string.gsub(s or "", "[}{\">/<'&]", {
            ["&"] = "&amp;",
            ["<"] = "&lt;",
            [">"] = "&gt;",
            ['"'] = "&quot;",
            ["'"] = "&#39;",
            ["{"] = "&#123;",
            ["}"] = "&#125;"
        }))
    end
    
    return 200, [[
        <h1>Node Editor</h1>
        <form id="nodeForm" onsubmit="handleSubmit(event)">
            <div style="margin-bottom: 1rem;">
                <label for="nodePath">Node Path:</label>
                <div style="display: flex; gap: 0.5rem;">
                    <input type="text" id="nodePath" name="nodePath" value="]] .. nodePath .. [[" required>
                    <button type="button" onclick="loadContent()" style="background-color: #2196F3;">Load</button>
                </div>
            </div>
            
            <div style="margin-bottom: 1rem;">
                <label for="content">Content:</label>
                <textarea id="content" name="content" rows="20" required>]] .. escapeHTML(content) .. [[</textarea>
            </div>
            
            <button type="submit">Save or Create</button>
        </form>
        
        <div id="response" style="display: none; margin-top: 1rem;"></div>

        <style>
            form {
                max-width: 100%;
                padding: 1rem;
                border: 1px solid #ddd;
                border-radius: 8px;
            }
            
            input, textarea, button {
                padding: 0.5rem;
                font-family: monospace;
                border: 1px solid #ddd;
                border-radius: 4px;
                box-sizing: border-box;
				margin-top: 1em;
            }
            
            input, textarea {
                width: 100%;
            }
            
            input:focus, textarea:focus {
                outline: none;
                border-color: #2196F3;
                box-shadow: 0 0 0 2px rgba(33, 150, 243, 0.2);
            }
            
            textarea {
                white-space: pre;
                overflow-wrap: normal;
                overflow-x: auto;
                min-height: 300px;
                line-height: 1.4;
                resize: vertical;
            }
			
			label {
				margin-bottom: 1rem;
			}
            
            button {
                background-color: #4CAF50;
                color: white;
                border: none;
                cursor: pointer;
                transition: background-color 0.3s ease;
                line-height: normal;
            }
            
            button:hover {
                background-color: #45a049;
            }

            .success-message {
                padding: 1rem;
                background-color: #dff0d8;
                border: 1px solid #d0e9c6;
                border-radius: 4px;
                color: #3c763d;
                margin-top: 1rem;
            }

            .error-message {
                padding: 1rem;
                background-color: #f2dede;
                border: 1px solid #ebcccc;
                border-radius: 4px;
                color: #a94442;
                margin-top: 1rem;
            }
        </style>

        <script>
        async function handleSubmit(event) {
            event.preventDefault();
            const form = event.target;
            const responseDiv = document.getElementById('response');
            const submitButton = form.querySelector('button[type="submit"]');
            
            submitButton.disabled = true;
            submitButton.textContent = 'Saving...';
            
            try {
                const formData = new FormData(form);
                const data = {};
                formData.forEach((value, key) => data[key] = value);
                
                const response = await fetch(window.location.href, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                responseDiv.style.display = 'block';
                if (result.status === 'success') {
                    responseDiv.innerHTML = `
                        <div class="success-message">
                            <h3>Save Successful!</h3>
                            <p>Node path: ${result.data.nodePath}</p>
                        </div>
                    `;
                    // Update URL with the node path
                    const url = new URL(window.location);
                    url.searchParams.set('path', data.nodePath);
                    window.history.pushState({}, '', url);
                } else {
                    throw new Error(result.message || 'Save failed');
                }
            } catch (error) {
                responseDiv.style.display = 'block';
                responseDiv.innerHTML = `
                    <div class="error-message">
                        <p>Error: ${error.message || 'Save failed, please try again'}</p>
                    </div>
                `;
            } finally {
                submitButton.disabled = false;
                submitButton.textContent = 'Save';
            }
        }

        async function loadContent() {
            const path = document.getElementById('nodePath').value;
            const responseDiv = document.getElementById('response');
            
            if (!path) {
                responseDiv.style.display = 'block';
                responseDiv.innerHTML = `
                    <div class="error-message">
                        <p>Error: Please enter a node path</p>
                    </div>
                `;
                return;
            }
            
            const url = new URL(window.location);
            url.searchParams.set('path', path);
            
            try {
                const response = await fetch(url);
                if (!response.ok) {
                    throw new Error('Failed to load content');
                }
                const text = await response.text();
                const contentStart = text.indexOf('<textarea');
                const contentEnd = text.indexOf('</textarea>');
                if (contentStart === -1 || contentEnd === -1) {
                    throw new Error('Invalid response format');
                }
                
                // If content is empty, it means the node doesn't exist
                const textareaContent = text.substring(contentStart, contentEnd + 11);
                const tempDiv = document.createElement('div');
                tempDiv.innerHTML = textareaContent;
                const content = tempDiv.querySelector('textarea').value;
                
                if (!content) {
                    responseDiv.style.display = 'block';
                    responseDiv.innerHTML = `
                        <div class="error-message">
                            <p>Error: Node not found</p>
                        </div>
                    `;
                    document.getElementById('content').value = '';
                } else {
                    window.history.pushState({}, '', url);
                    document.getElementById('content').value = content;
                    responseDiv.style.display = 'none';
                }
            } catch (error) {
                responseDiv.style.display = 'block';
                responseDiv.innerHTML = `
                    <div class="error-message">
                        <p>Error: ${error.message}</p>
                    </div>
                `;
            }
        }
        </script>
    ]]
end

function post(request)
    local nodePath = request.params["nodePath"]
    local content = request.params["content"]
    
    if not nodePath or nodePath == "" then
        return 400, '{"status": "error", "message": "Node path is required"}'
    end
    
    if not content then
        return 400, '{"status": "error", "message": "Content is required"}'
    end
    
    local success = createNode(nodePath, content)
    if not success then
        return 500, '{"status": "error", "message": "Failed to save node"}'
    end
    
    return 200, string.format([[
        {
            "status": "success",
            "message": "Node saved successfully",
            "data": {
                "nodePath": "%s"
            }
        }
    ]], nodePath)
end

function put(request)
    -- PUT requests will be returned directly
    return 200, '{"status": "updated", "message": "Data updated"}'
end
