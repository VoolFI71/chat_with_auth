function checkEnter(event) {
    if (event.key === 'Enter') {
        event.preventDefault(); 
        createMessage();
    }
}


function logout() {
    localStorage.removeItem('token');
    window.location.reload();
}

const token = localStorage.getItem('token');

if (token) {
    document.getElementById('auth-buttons').style.display = 'none';
    document.getElementById('user-info').style.display = 'block';
    getUserInfo();
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
        document.getElementById('username-display').textContent = data.username;
    })
    .catch(error => {
        console.error('Error fetching user info:', error);
    });
}

let id = 0;

function getMessages() {
    const url = `http://127.0.0.1:8080/getmsg?id=${id}`;
    console.log(id)
    fetch(url, {
        method: 'GET',
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        console.log(data.id)

        const messagesList = document.getElementById('messages'); 
        messagesList.innerHTML = ''; 
        const fragment = document.createDocumentFragment(); 

        data.messages.forEach(msg => {
            const li = document.createElement('li');
        
            const textNode = document.createElement('span');
            textNode.textContent = `${msg.username}: `;
            li.appendChild(textNode); 
        
            if (msg.message) {
                const messageText = document.createElement('span');
                messageText.textContent = msg.message;
                li.appendChild(messageText); 
            }
        
            if (msg.image) {
                const img = document.createElement('img');
                img.src = msg.image;
                img.style.maxWidth = '200px';
                img.style.display = 'block';
                li.appendChild(img); 
            }
        
            if (msg.audio) {
                const audio = document.createElement('audio');
                audio.controls = true; 
                audio.src = msg.audio; 
                li.appendChild(audio); 
            }

            fragment.appendChild(li);
        });
        messagesList.append(fragment); 
        console.log('New ID from server:', data.id);
        id = data.id || 0;
    })
    .catch(error => {
        console.error('Error fetching messages:', error);
    });
}



const messagesContainer = document.getElementById('messages');
messagesContainer.addEventListener('scroll', () => {
    if (messagesContainer.scrollTop === 0) {
        getMessages(); 
    }
});

window.onload = function() {
    getMessages(); 
};


const conn = new WebSocket(`ws://127.0.0.1:8080/ws`);

const messagesList = document.getElementById('messages');

conn.onmessage = function(event) {
    const data = JSON.parse(event.data);
    const li = document.createElement('li');
    if (data.message) {
        li.textContent = `${data.username}: ${data.message}`; 
    }

    if (data.image) {
        const img = document.createElement('img');
        img.src = data.image;
        img.style.maxWidth = '200px';
        img.style.display = 'block'; 
        li.textContent = `${data.username}:`
        li.appendChild(img);
    }
    if (data.audio) {
        const audio = document.createElement('audio');
        audio.controls = true; 
        audio.src = `data:audio/wav;base64,${data.audio}`;
        li.textContent = `${data.username}:`;
        li.appendChild(audio);
    }

    document.getElementById('messages').prepend(li);
};


function createMessage() {
    const messageInput = document.getElementById('message');
    const message = messageInput.value.trim();
    
    if (message) {
        const messageData = { message: message };

        fetch('http://127.0.0.1:8080/savemsg', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(messageData) 
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
            conn.send(JSON.stringify(msg)); 
            document.getElementById('message').value = ''; 
            })
            .catch(error => {
                console.error('Ошибка:', error);
        });
    }
}
    
function createImage() {
    const imageInput = document.getElementById('image')
    const image = imageInput.files.length > 0 ? imageInput.files[0] : null;
    const formData = new FormData(); 

    formData.append('image', image);
    if (image) {
        fetch('http://127.0.0.1:8080/saveimage', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
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
            const msg = { username: data.username, image: image };

            const reader = new FileReader();
            reader.onload = function(event) {
                const base64Image = event.target.result; // Получаем Base64 строку
                msg.image = base64Image; // Добавляем изображение в сообщение
                conn.send(JSON.stringify(msg)); // Отправляем сообщение по WebSocket
            };
            reader.readAsDataURL(image);
            imageInput.value = '';
            })
            .catch(error => {
                console.error('Ошибка:', error);
        })
    }
}

function checktype() {
    const imageInput = document.getElementById('image')
    if (imageInput){
        const image = imageInput.files.length > 0 ? imageInput.files[0] : null;
    }
    const message = document.getElementById('message').value.trim();
    if (message){
        createMessage()
    }
    if (image){
        createImage()
    }
}


let mediaRecorder;
let audioChunks = [];
let isRecording = false;

function toggleRecording() {
    if (isRecording) {
        mediaRecorder.stop();
        isRecording = false;
        document.getElementById('recordButton').innerText = 'Записать аудио';
    } else {
        startRecording();
        isRecording = true;
        document.getElementById('recordButton').innerText = 'Остановить запись';
    }
}

async function startRecording() {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    mediaRecorder = new MediaRecorder(stream);
    let audioChunks = []; // Объявляем переменную для хранения аудиочастей

    mediaRecorder.ondataavailable = event => {
        audioChunks.push(event.data);
    };

    mediaRecorder.onstop = async () => {
        const audioBlob = new Blob(audioChunks, { type: 'audio/wav' });
        audioChunks = [];
        const audioUrl = URL.createObjectURL(audioBlob);

        const formData = new FormData();
        formData.append('audio', audioBlob, 'audio.wav'); // Добавляем аудиофайл в FormData

        fetch('http://127.0.0.1:8080/saveaudio', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },  
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
            const msg = { username: data.username, audio: 1 }; 
            
            const reader = new FileReader();
            reader.onload = function(event) {
                const base64Audio = event.target.result.split(',')[1]; // Получаем только Base64 часть
                msg.audio = base64Audio; // Добавляем аудиоданные в сообщение
                conn.send(JSON.stringify(msg)); // Отправляем сообщение по WebSocket
            };
            reader.readAsDataURL(audioBlob); 
        })
        .catch(error => {
            console.error('Ошибка сети:', error);
        });
    };

    mediaRecorder.start();
}