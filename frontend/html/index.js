function logout() {
    localStorage.removeItem('token');
    window.location.reload();
    }

    const token = localStorage.getItem('token');

    if (token) {
        document.getElementById('auth-buttons').style.display = 'none';
        document.getElementById('user-info').style.display = 'block';
        getUserInfo();
        getMessages();
    } else {
        document.getElementById('auth-buttons').style.display = 'block';
        document.getElementById('user-info').style.display = 'none';
    }

    function redirectToRegister() {
        window.location.href = 'http://127.0.0.1/reg'; 
    }

    function redirectToLogin() {
        window.location.href = 'http://127.0.0.1/login'; 
    }

    function getUserInfo() {
        fetch('http://127.0.0.1:8080/userinfo', {
            method: 'GET',
            credentials: 'include',

            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            document.getElementById('username-display').textContent = `Привет, ${data.username}!`;
        })
        .catch(error => {
            console.error('Error fetching user info:', error);
        });
    }

    function getMessages() {
        fetch('http://127.0.0.1:8080/getmsg', {
            method: 'GET', 
            // headers: {
            //     'Content-Type': 'application/json',
            // }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            const messagesList = document.getElementById('messages'); 
            messagesList.innerHTML = ''; 
            const fragment = document.createDocumentFragment(); // Создаем DocumentFragment

            data.forEach(msg => {
                const li = document.createElement('li');
                li.textContent = `${msg.username}: ${msg.message}`;

                if (msg.image) {
                    const img = document.createElement('img');
                    img.src = msg.image; // Устанавливаем src на строку Base64
                    img.src = `data:image/png;base64,${msg.image}`;

                    img.style.maxWidth = '200px'; // Ограничиваем размер изображения
                    img.style.display = 'block'; // Отображаем изображение как блок
                    li.appendChild(img); //
                }

                fragment.appendChild(li); // Добавляем элемент li во фрагмент
            });
            messagesList.appendChild(fragment); // Добавляем все элементы во фрагмент за один раз

        })
        .catch(error => {
            console.error('Error fetching messages:', error);
        });
    }

    window.onload = function() {
        getMessages(); 
    };


    const conn = new WebSocket(`ws://127.0.0.1:8080/ws`);

    const messagesList = document.getElementById('messages');
    
    conn.onmessage = function(event) {
        const msg = JSON.parse(event.data);
        const li = document.createElement('li');

        li.textContent = `${msg.username}: ${msg.message}`; 

        if (msg.image) {
            const img = document.createElement('img');
            img.src = msg.image; // Устанавливаем src на строку Base64
            img.style.maxWidth = '200px'; // Ограничиваем размер изображения
            img.style.display = 'block'; // Отображаем изображение как блок
            li.appendChild(img); //
        }

        messagesList.prepend(li);
    };

    function checkEnter(event) {
        if (event.key === 'Enter') {
            event.preventDefault(); 
            createMessage();
        }
    }

        function createMessage() {
            const message = document.getElementById('message').value.trim();
            const imageInput = document.getElementById('image')
            const image = imageInput.files.length > 0 ? imageInput.files[0] : null;
            const formData = new FormData(); 
            if (!message && !image) {
                alert('Пожалуйста, введите сообщение или загрузите изображение.');
                return;
            }

            if (message) {
                formData.append('message', message);
            }
            if (image) {
                formData.append('image', image); 
            }


            if (message || image) {
                fetch('http://127.0.0.1:8080/savemsg', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`
                    },
                    // body: JSON.stringify({
                    //     message: message   
                    // })
                    body: formData
                })
            .then(response => {
                if (response.ok) {
                    return response.json();
                } else if (response.status === 401) { 
                    alert('Необходимо авторизоваться');
                    throw new Error('Необходимо авторизоваться');
                } else if (response.status === 400) {
                    alert('Некорректный запрос');
                    throw new Error('Некорректный запрос');
                } else {
                    alert('Произошла ошибка: ' + response.status);
                    throw new Error('Произошла ошибка: ' + response.status);
                }
            })
            .then(data => {
                const msg = { username: data.username, message: message };
                // Если есть изображение, преобразуем его в Base64
                if (image) {
                    const reader = new FileReader();
                    reader.onload = function(event) {
                        const base64Image = event.target.result; // Получаем Base64 строку
                        msg.image = base64Image; // Добавляем изображение в сообщение
                        conn.send(JSON.stringify(msg)); // Отправляем сообщение по WebSocket
                    };
                    reader.readAsDataURL(image); // Читаем изображение как Data URL
                } else {
                    conn.send(JSON.stringify(msg)); 
                }

                document.getElementById('message').value = ''; 
                imageInput.value = '';
            })
            .catch(error => {
                console.error('Ошибка:', error);
            });
        }
    }