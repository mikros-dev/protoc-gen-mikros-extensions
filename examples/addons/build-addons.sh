#!/bin/bash

for addon in *; do
    if ! [ -d "$addon" ]; then
        continue
    fi

    echo "Building addon $addon"

    # compile protos move them to the plugin mikros/extensions directory and
    # copy the protos to protobuf/addons
    count=`ls -1 $addon/*.proto 2>/dev/null | wc -l`
    if [ $count != 0 ]; then
        # compile the proto
        (cd $addon &&   \
          make clean && \
          make proto && \
          cp -f *.pb.go ../../../mikros/extensions)

        # copy proto file for examples
        cp -f $addon/*.proto ../protobuf/addons
    fi

    # compile the addons and move them to the protobuf/addons directory
    (cd $addon &&   \
      make clean && \
      make &&       \
      cp -f *.so ../../protobuf/addons)

    echo ""
done

exit 0