'use strict';

var WebSocketHandler = {
  isConnected: false,
  webSocket: null,
  connect: function() {
    var ws = new WebSocket("ws://localhost:8080/play/");
    WebSocketHandler.webSocket = ws;
    ws.onopen = function() {
      console.log("WebSocket opened");
      WebSocketHandler.isConnected = true;
      ws.onclose = function() {
        console.log("WebSocket closed");
        WebSocketHandler.ws = null;
        WebSocketHandler.isConnected = false;
      }
    }
  },
  send: function(dataToSend, receiveCallback) {
    // check if connected
    if (WebSocketHandler.isConnected) {
      // send stuff
      WebSocketHandler.webSocket.send(JSON.stringify(dataToSend));
      // set new callback on ws.onmessage -> receiveCallbackFunctionPointer
      WebSocketHandler.webSocket.onmessage = function(message) {
        receiveCallback(JSON.parse(message.data));
      };
    } else {
      console.log("Error: Not yet connected.");
    }
  },
  listen: function(receiveCallback) {
    WebSocketHandler.webSocket.onmessage = function(message) {
      receiveCallback(JSON.parse(message.data));
    };
  }
};

WebSocketHandler.connect({});


/* Controllers */

function AppCtrl($scope, $q, $rootScope) {

$(function () {
    $('#info').popover({'html': true});

});

// For the timer
$scope.roundProgressData = {
      label: 0,
      percentage: 0
}

$scope.signalIcons = ['fa-times-circle', 'fa-check-circle', 'fa-smile-o', 'fa-frown-o'];
$scope.signals = [];
$scope.join = false;
$scope.authenticated = false;
$scope.matched = false;
$scope.round = 1;
$scope.balance = 0;
//$scope.authenticated = true;
//$scope.matched = true;
//$scope.join = true;

    // Keep all pending requests here until they get responses
    var callbacks = {};

    // Create a unique callback ID to map requests to responses
    var currentCallbackId = 0;

    // Create our websocket object with the address to the websocket
    var ws = new WebSocket("ws://localhost:8080/play/");
  
    ws.onmessage = function(message) {
        console.log(message.data);
        listener(JSON.parse(message.data));
    };

    function sendRequest(request) {
      var defer = $q.defer();
      var callbackId = getCallbackId();
      callbacks[callbackId] = {
        time: new Date(),
        cb:defer
      };

      request.callback_id = callbackId;
      console.log('Sending request', request);
      ws.send(JSON.stringify(request));
      return defer.promise;
    }

    function listener(data) {
      var messageObj = data;

      console.log("Received data from websocket: ", messageObj);
      if (typeof messageObj.result.errorCode != "undefined") {
        console.log('Error: ' + messageObj.result.errorMsg);
      }

      if(messageObj.command == "matched") {
        $scope.matched = true;
      }

      if(messageObj.command == "login" && messageObj.result == "success") {
        console.log("Succesfully logged in!")
        $scope.authenticated = true;
        console.log($scope.authenticated)
      }

      // If an object exists with callback_id in our callbacks object, resolve it
      if(callbacks.hasOwnProperty(messageObj.callback_id)) {
        console.log(callbacks[messageObj.callback_id]);
        $rootScope.$apply(callbacks[messageObj.callback_id].cb.resolve(messageObj.data));
        delete callbacks[messageObj.callbackID];
      }
    }
    // This creates a new callback ID for a request
    function getCallbackId() {
      currentCallbackId += 1;
      if(currentCallbackId > 10000) {
        currentCallbackId = 0;
      }
      return currentCallbackId;
    }

$scope.getBalance = function() {
  WebSocketHandler.send({'command': 'getBalance'}, function(data) {
    if(data.command == 'balance') {
      console.log(data)
      console.log('balance: ' + data.result);
      $scope.balance = data.result;
    }
  });
};

$scope.login = function(name, password) {
  WebSocketHandler.send({command: 'login', name: name, password: password}, function(data) {
    if(data.result == 'success') {
      $scope.authenticated = true;
      $scope.getBalance();
    }
  });

};

$scope.register = function(name, password) {
  WebSocketHandler.send({command: 'register', name: name, password: password}, function(data) {
    if(data.result == 'success') {
      $scope.authenticated = true;
    }
  });
};


// Player indicates he wants to start a new game
$scope.joinGame = function() {
  WebSocketHandler.send({'command': 'join'});
  $scope.join = true;
}

$scope.sendRequestOnOpen = function(request) {
  return sendRequest(request);
}

$scope.$watch('roundProgressData', function (newValue, oldValue) {
  newValue.percentage = newValue.label / 100;
}, true);

/*
socket.on('')

// Initiate the game (set balance for user)
socket.on('init', function (data) {
  socket.emit('get:balance')
});

// The user gets the requested balance of the server
socket.on('balance', function (data) {
  $scope.balance = data.result;
});

// The user gets the requested depositaddress of the server
socket.on('deposit:address', function (data) {
  $scope.depositAddress = data.result;
});

// The user recieves a message from the other user
socket.on('send:signal', function (data) {
  $scope.signals.push({
    player: 'oponent',
    signal: data.signal
  });
});

// Let the games begin
socket.on('matched', function (data) {
  $scope.matched = true;
});

// The user gets a confirmation/error on withdraw
socket.on('withdraw', function (data) {
  $scope.withdraw = data.result;
});

// Other player played
socket.on('outcome', function (data) {
  socket.emit('get:balance')
  $scope.outcome = data.result;
});

// Methods published to the scope
// ==============================

// Request a depost address
$scope.getDepositAddress = function() {
  socket.emit('get:deposit:address')
}

// The player has made a decision 
$scope.sendAction = function(action) {
  socket.emit('command:action', {
     action: action
  });

  console.log('Action: ' + action)
  console.log('Join: ' + $scope.join)
  console.log('Matched: ' + $scope.matched)
};

// The player sends a signal to the other player

$scope.countDown = function(seconds) {
  $scope.roundProgressData.percentage = 0
  $scope.roundProgressData.label = seconds
  for (var i = 1; i <= seconds; i++) {
    $scope.roundProgressData.percentage += 100.0/seconds ;
    $scope.roundProgressData.label = seconds - i;

    console.log($scope.roundProgressData)
  }
}

*/

$scope.sendSignal = function(signal) {
  $scope.signals.push({'player': 'you', 'signal': signal});
  WebSocketHandler.send({command: 'signal', signal: signal}, function(data) {
    console.log(data);
  });
}

$scope.$watch(function() {
  return $('.popover.fade.in').attr('opacity'); 
}, function(newValue){
  if (newValue == 0) {
    console.log('dissapear')
  }

});

$scope.$watchCollection('signals', function() {
  console.log($('div.signaloverview').scrolltop);
  $(".signaloverview").animate({ scrollTop: $('.signaloverview').height()}, 1000);
});

}
