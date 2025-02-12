function render(request)
    return 200, [[
        <h1>User Information Submission</h1>
        <form id="userForm" onsubmit="handleSubmit(event)">
            <div style="margin-bottom: 1rem;">
                <label for="name">Name:</label>
                <input type="text" id="name" name="name" required>
            </div>
            
            <div style="margin-bottom: 1rem;">
                <label for="email">Email:</label>
                <input type="email" id="email" name="email" required>
            </div>
            
            <div style="margin-bottom: 1rem;">
                <label for="message">Message:</label>
                <textarea id="message" name="message" rows="4" required></textarea>
            </div>
            
            <button type="submit">Submit</button>
        </form>
        
        <div id="response" style="display: none; margin-top: 1rem;"></div>

        <style>
            form {
                max-width: 500px;
                margin: 2rem auto;
                padding: 1rem;
                border: 1px solid #ddd;
                border-radius: 8px;
            }
            
            input, textarea {
                width: 100%;
                padding: 0.5rem;
                margin-top: 0.25rem;
            }
            
            button {
                background-color: #4CAF50;
                color: white;
                padding: 0.5rem 1rem;
                border: none;
                border-radius: 4px;
                cursor: pointer;
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
            
            // Disable submit button to prevent duplicate submissions
            submitButton.disabled = true;
            submitButton.textContent = 'Submitting...';
            
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
                            <h3>Submission Successful!</h3>
                            <p>Name: ${result.data.name}</p>
                            <p>Email: ${result.data.email}</p>
                            <p>Message: ${result.data.message}</p>
                        </div>
                    `;
                    form.reset();
                } else {
                    throw new Error(result.message || 'Submission failed');
                }
            } catch (error) {
                responseDiv.style.display = 'block';
                responseDiv.innerHTML = `
                    <div class="error-message">
                        <p>Error: ${error.message || 'Submission failed, please try again'}</p>
                    </div>
                `;
            } finally {
                // Restore submit button
                submitButton.disabled = false;
                submitButton.textContent = 'Submit';
            }
        }
        </script>
    ]]
end

function post(request)
    -- Get JSON data from POST request body
    local name = request.params["name"] or ""
    local email = request.params["email"] or ""
    local message = request.params["message"] or ""
    
    -- Add data validation and processing logic here
    
    -- Return JSON response
    return 200, string.format([[
        {
            "status": "success",
            "message": "Form submitted successfully",
            "data": {
                "name": "%s",
                "email": "%s",
                "message": "%s"
            }
        }
    ]], name, email, message)
end

function put(request)
    -- PUT requests will be returned directly
    return 200, '{"status": "updated", "message": "Data updated"}'
end