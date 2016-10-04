var source = new window.EventSource('/events');
source.onmessage = function (e) {
    document.body.innerHTML += e.data + '<br />';
};
