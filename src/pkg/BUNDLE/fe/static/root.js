
function liveOK(data) {
        $('#live').html(data.Links)
}

function liveError(data) {
        $('#live').html('<p>We lost connection to the Tonika program. Is it still running?</p>')
}

function refresh() {
        $.ajax({
                url: '/api/live',
                success: liveOK,
                error: liveError,
                dataType: 'json',
        });
}

$(document).ready(function(){
        mainReady();
        window.setInterval('refresh()', 1000)
});
