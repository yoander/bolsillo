$(document).ready(function () {
    $('#filter').click(function() {
        document.body.style.cursor='wait';
        toastr.options = {
            "closeButton": false,
            "debug": false,
            "newestOnTop": false,
            "progressBar": false,
            "positionClass": "toast-bottom-right",
            "preventDuplicates": false,
            "onclick": null,
            "showDuration": "300",
            "hideDuration": "1000",
            "timeOut": "0",
            "extendedTimeOut": "0",
            "showEasing": "swing",
            "hideEasing": "linear",
            "showMethod": "fadeIn",
            "hideMethod": "fadeOut"
        }

        toastr.info("Loading transactions");

        var filter = $(this);
        $.ajax({
            url: $(this).attr('data-url'),
            data: {
               startDate: $('#startDate').val(),
               endDate: $('#endDate').val(),
               keyword: $('#description').val(),
            },
            beforeSend: function() {
                filter.prop('disabled', true);
            },
            success: function(data) {
                toastr.success("Transactions loaded");
                $('#transTable').html(data);
            },
            complete: function () {
                window.setTimeout(function() { toastr.clear(); }, 3000)
                filter.prop('disabled', false);
                document.body.style.cursor='default';
            }
        });
    }) ;
});