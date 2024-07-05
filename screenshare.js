const startShareButton = document.getElementById('startShare');
let mediaRecorder;
let socket;

const boundary = '--screenshare-boundary';

startShareButton.addEventListener('click', async () => {
    try {
        const stream = await navigator.mediaDevices.getDisplayMedia({
            video: true,
            audio: false
        });

        const videoTrack = stream.getVideoTracks()[0];
        const mediaStream = new MediaStream([videoTrack]);

        mediaRecorder = new MediaRecorder(mediaStream, { mimeType: 'video/mp4' });
        socket = new WebSocket('ws://localhost:8069/screenshare');

        mediaRecorder.ondataavailable = (event) => {
            socket.send(event.data, { binary: true });
        };

        mediaRecorder.start();
    } catch (err) {
        console.error('Error accessing screen:', err);
    }
});
