$(document).ready(function () {
    const videoElement = $('#videoPlayer')[0];
    const wsUrl = 'ws://localhost:8080/predict';

    function isValidUrl() {
        return /^(http|https|ftp):\/\/[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$/i.test($("#url").val());
    }

    $('#predictButton').on('click', function() {
        if (!isValidUrl()) {
            alert($('#url').val()+'is not valid');
            return
        }
        const socket = new WebSocket(wsUrl+'?url='+$('#url').val());
        let playingvideo = false

        socket.binaryType = 'arraybuffer';

        socket.onopen = function (event) {
            console.log('WebSocket connection opened:', event);
        };

        socket.onmessage = function (event) {
            console.log("received message: " + event.data)
            if (!playingvideo) {
                $('#logs').text(event.data + '\n' + $('#logs').text());

                if (event.data === 'DONE.') {
                    playingvideo = true;
                }
                return
            }
            // Handle incoming binary data
            // const blob = new Blob([event.data], { type: 'video/mp4' });
            const videoUrl = 'http://localhost:8080'+event.data
            if (Hls.isSupported()) {
                var hls = new Hls({
                    debug: true,
                });

                hls.loadSource(videoUrl);
                hls.attachMedia(videoElement);
                hls.on(Hls.Events.MEDIA_ATTACHED, function () {
                    videoElement.muted = true;
                    videoElement.play();
                });
            } else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
                console.log('here')
                videoElement.src = videoUrl;
                videoElement.addEventListener('canplay', function () {
                    videoElement.play();
                });
            }
        };

        socket.onclose = function (event) {
            console.log('WebSocket connection closed:', event);
        };

        socket.onerror = function (error) {
            console.error('WebSocket error:', error);
        };
    });
});