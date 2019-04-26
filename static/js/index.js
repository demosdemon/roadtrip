let map;

// noinspection JSUnusedGlobalSymbols
function initMap() {
    map = new google.maps.Map(document.getElementById("map"), {
        center: {
            lat: 48.868297,
            lng: 2.353764,
        },
        zoom: 6,
    });

    if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(function (position) {
            const pos = {
                lat: position.coords.latitude,
                lng: position.coords.longitude,
            };

            map.setCenter(pos);
        });
    }
}
