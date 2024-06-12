document.getElementById('startStopButton').addEventListener('click', toggleRecording);

let mediaRecorder;
let webSocket;
let isRecording = false;

async function toggleRecording() {
    if (isRecording) {
        stopRecording();
    } else {
        startRecording();
    }
}

async function startRecording() {
    try {
        const stream = await navigator.mediaDevices.getDisplayMedia({ video: true });
        mediaRecorder = new MediaRecorder(stream, { mimeType: 'video/webm' });

        webSocket = new WebSocket('wss://yourserver.com/path');

        webSocket.onopen = () => {
            console.log('WebSocket connection opened');
            mediaRecorder.start(1000); // Send data in chunks every second

            mediaRecorder.ondataavailable = (event) => {
                if (event.data.size > 0 && webSocket.readyState === WebSocket.OPEN) {
                    webSocket.send(event.data);
                }
            };

            mediaRecorder.onstop = () => {
                stream.getTracks().forEach(track => track.stop());
                webSocket.close();
                console.log('Recording stopped');
            };

            isRecording = true;
            document.getElementById('startStopButton').innerText = 'Stop Recording';
        };

        webSocket.onerror = (error) => {
            console.error('WebSocket error:', error);
            stopRecording();
        };

        webSocket.onclose = () => {
            console.log('WebSocket connection closed');
            stopRecording();
        };
    } catch (error) {
        console.error('Error accessing display media:', error);
    }
}

function stopRecording() {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
        mediaRecorder.stop();
    }
    isRecording = false;
    document.getElementById('startStopButton').innerText = 'Start Recording';
}

