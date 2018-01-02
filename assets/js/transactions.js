$(document).ready(function () {
    $('#filter').click(function() {
        var filter = $(this);
        filter.prop('disabled', true);
        var n = new Noty({
            type: 'info',
            layout: 'bottomRight',
            theme: 'bootstrap-v4',
            text: '<strong><i class="fa fa-spinner fa-spin fa-fw"></i> Loading transactions</strong>'
        });

        n.show();

        $.ajax({
            url: $(this).attr('data-url'),
            data: {
               startDate: $('#startDate').val(),
               endDate: $('#endDate').val(),
            },
            success: function(data) {
                console.log('OK')
                console.log(data);
                $('#transTable').html(data);
                n.setType('success'); // Notification type updater
                n.setText('<strong><i class="fa fa-check" aria-hidden="true"></i> Transactions loaded</strong>');
            },
            complete: function () {
                n.setTimeout(500);
                filter.prop('disabled', false);
            }
        });
        return false;
    }) ;
});