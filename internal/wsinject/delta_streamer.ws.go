package wsinject

const DeltaStreamerSourceCode = `/**
* This file has been injected by the wd-41 web development
* hot reload tool. 
*/

function startWebsocket() {
  // Check if the WebSocket object is available in the current context
  if (typeof WebSocket !== 'function') {
    console.error('WebSocket is not supported by this browser.');
    return;
  }

  // Establish a connection with the WebSocket server
  const socket = new WebSocket('ws://localhost:%v%v');

  // Event handler for when the WebSocket connection is established
  socket.addEventListener('open', function (event) {
    console.log('Connected to the WebSocket server');
  });

  // Event handler for when a message is received from the server
  socket.addEventListener('message', function (event) {
    console.log('Message from server:', event.data);
    let fileName = window.location.pathname.split('/').pop();
    if(fileName === "") {
      fileName = "/index.html"
    } else {
      fileName = "/" + fileName
    }
    // Reload page if it's detected that the current page has been altered
    if(event.data === fileName ||
      // Always reload on js and css files since its difficult to know where these are used
      event.data.includes(".js") ||
      event.data.includes(".css")) {
      location.reload();
    }
  });

  // Event handler for when the WebSocket connection is closed
  socket.addEventListener('close', function (event) {
    console.log('Disconnected from the WebSocket server');
  });

  // Event handler for when an error occurs with the WebSocket connection
  socket.addEventListener('error', function (event) {
    console.error('WebSocket error:', event);
    console.error(event.message)
  });
}

startWebsocket();
`
