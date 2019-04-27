const root = {
    lat: 48.868297,
    lng: 2.353764,
};

let map, marker;
let markerDrag = false;

function updateMarker() {
    if (!markerDrag) {
        const pos = marker.getPosition();
        console.log("position_changed");
        console.log(`lat: ${pos.lat()}, lng: ${pos.lng()}`);
        // TODO: send position to backend

        $.ajax({
            url: "/location",
            method: "POST",
            data: {
                lat: pos.lat(),
                lng: pos.lng(),
            },
        }).done(function (data) {
            console.log("result");
            console.dir(data);
        }).fail(function (e) {
            console.log("error");
            console.dir(e);
        });
    }
}

function plopMarker(pos) {
    marker = new google.maps.Marker({
        animation: google.maps.Animation.DROP,
        draggable: true,
        position: pos,
        map: map,
    });

    marker.addListener("dragstart", function () {
        markerDrag = true;
    });

    marker.addListener("dragend", function () {
        markerDrag = false;
        updateMarker();
    });

    marker.addListener("position_changed", updateMarker);

    updateMarker();
}

// noinspection JSUnusedGlobalSymbols
function initMap() {
    map = new google.maps.Map(document.getElementById("map"), {
        center: root,
        zoom: 6,
        fullscreenControl: false,
        gestureHandling: "cooperative",
    });

    map.addListener("dblclick", function (ev) {
        marker.setPosition(ev.latLng);
    });

    $("#btn-locate-me").click(function () {
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(geolocationCallback, geolocationErrorCallback);
        } else {
            handleLocationError(false, map.getCenter());
        }
    });

    plopMarker(root);
}

function handleLocationError(browserHasGeolocation, pos) {
    const infoWindow = new google.maps.InfoWindow();
    infoWindow.setPosition(pos);
    if (browserHasGeolocation) {
        infoWindow.setContent("Error: the geolocation service failed.");
    } else {
        infoWindow.setContent("Error: your browser doesn't support geolocation.")
    }
    infoWindow.open(map);
}

function geolocationCallback(position) {
    const pos = {
        lat: position.coords.latitude,
        lng: position.coords.longitude,
    };
    map.panTo(pos);
    marker.setPosition(pos);
}

function geolocationErrorCallback(e) {
    console.dir(e);
    handleLocationError(true, map.getCenter());
}
